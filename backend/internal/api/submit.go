package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"gotutor/backend/chapters"
	"gotutor/backend/internal/verifier"
)

// maxBodyBytes caps the request body so a giant payload can't OOM us
// before the verifier runs. 256 KiB is plenty for any realistic Go file.
const maxBodyBytes = 256 * 1024

// submitTimeout is the outer deadline for the whole submit handler.
// The verifier adds an inner per-test timeout via Policy.GoTestTimeout.
const submitTimeout = 15 * time.Second

// HandleSubmit — POST /api/chapters/:id/submit
// Body: {"userCode": "package main ..."}
// Response: SubmitResult JSON with passed/output/durationMs/nextChapterUnlocked.
//
// Flow:
//  1. Look up chapter; reject 404 if unknown.
//  2. Bind + cap the request body (256 KiB).
//  3. Run verifier.Verify with a 15s outer ctx.
//  4. On pass: upsert progress with completed_at; check if next chapter newly unlocked.
//  5. On fail: increment attempts but don't mark completed.
func HandleSubmit(c *gin.Context, db *sql.DB, goBin string) {
	id := c.Param("id")
	ch, ok := chapters.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown chapter"})
		return
	}

	// Cap body before JSON-decoding to bound memory.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body: " + err.Error()})
		return
	}

	var req SubmitRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}
	if req.UserCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userCode is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), submitTimeout)
	defer cancel()

	result := verifier.Verify(ctx, ch, req.UserCode, goBin)

	wasCompleted, _ := isCompleted(db, ch.ID)
	wasNextUnlocked, _ := isNextUnlocked(db, ch)

	if result.Passed {
		if err := markCompleted(db, ch.ID, result.Output); err != nil {
			// Even on write failure we still report pass — the user
			// succeeded; persistence is best-effort.
			result.Output += "\n[warning: failed to persist progress: " + err.Error() + "]"
		}
	} else {
		if err := incrementAttempts(db, ch.ID, result.Output); err != nil {
			_ = err
		}
	}

	nowNextUnlocked, _ := isNextUnlocked(db, ch)
	nextUnlocked := !wasNextUnlocked && nowNextUnlocked && !wasCompleted

	c.JSON(http.StatusOK, SubmitResult{
		Passed:              result.Passed,
		Output:              result.Output,
		DurationMs:          result.DurationMs,
		NextChapterUnlocked: nextUnlocked,
	})
}

// isCompleted returns true iff progress.completed_at is non-null.
func isCompleted(db *sql.DB, chapterID string) (bool, error) {
	var completedAt sql.NullInt64
	err := db.QueryRow(`SELECT completed_at FROM progress WHERE chapter_id = ?`, chapterID).Scan(&completedAt)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return completedAt.Valid, nil
}

// isNextUnlocked returns true iff the chapter after ch (by ordinal) is
// already unlocked, i.e. ch is already completed. Used to detect the
// transition edge so SubmitResult.NextChapterUnlocked fires once.
func isNextUnlocked(db *sql.DB, ch chapters.Chapter) (bool, error) {
	all := chapters.List()
	for i, c := range all {
		if c.ID == ch.ID && i+1 < len(all) {
			next := all[i+1]
			return isCompleted(db, next.ID)
		}
	}
	return false, nil
}

// markCompleted upserts the progress row with completed_at = now and
// last_output = result. Attempts keeps accumulating.
func markCompleted(db *sql.DB, chapterID, output string) error {
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO progress (chapter_id, completed_at, attempts, last_output)
		 VALUES (?, ?, COALESCE((SELECT attempts FROM progress WHERE chapter_id = ?), 0) + 1, ?)
		 ON CONFLICT(chapter_id) DO UPDATE SET completed_at = excluded.completed_at,
		                                       attempts = excluded.attempts,
		                                       last_output = excluded.last_output`,
		chapterID, now, chapterID, output,
	)
	if err != nil {
		return fmt.Errorf("upsert progress: %w", err)
	}
	return nil
}

// incrementAttempts bumps the attempts counter for a failed submission
// without touching completed_at. Creates the row if missing.
func incrementAttempts(db *sql.DB, chapterID, output string) error {
	_, err := db.Exec(
		`INSERT INTO progress (chapter_id, completed_at, attempts, last_output)
		 VALUES (?, NULL, 1, ?)
		 ON CONFLICT(chapter_id) DO UPDATE SET attempts = attempts + 1,
		                                       last_output = excluded.last_output`,
		chapterID, output,
	)
	if err != nil {
		return fmt.Errorf("increment attempts: %w", err)
	}
	return nil
}
