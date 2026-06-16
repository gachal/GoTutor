// Package api implements the HTTP handlers for the GoTutor REST API.
// Handlers live as free functions that take *gin.Context and *sql.DB;
// server/routes.go wires them with the server's DB so they can run
// progress queries without dragging the sql package through every layer.
package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"gotutor/backend/chapters"
)

// listCompleted returns a set of chapter IDs the user has completed.
// Used to compute Chapter.Completed and Chapter.Unlocked in list responses.
func listCompleted(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT chapter_id FROM progress WHERE completed_at IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// preferZh parses Accept-Language and returns true if Chinese is preferred.
// Default is English when the header is absent or ambiguous.
func preferZh(c *gin.Context) bool {
	h := strings.ToLower(c.GetHeader("Accept-Language"))
	if strings.HasPrefix(h, "zh") {
		return true
	}
	return strings.Contains(h, "zh")
}

// HandleListChapters — GET /api/chapters
// Returns all chapters with their completion state. All chapters are
// unlocked from the start — the tutor shows everything up front rather
// than gating progression, so learners can explore freely. `Completed`
// still reflects whether the user has ever passed the chapter.
func HandleListChapters(c *gin.Context, db *sql.DB) {
	all := chapters.List()
	completed, err := listCompleted(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query progress: " + err.Error()})
		return
	}

	zh := preferZh(c)
	out := make([]Chapter, 0, len(all))
	for _, ch := range all {
		isDone := completed[ch.ID]
		out = append(out, Chapter{
			ID:          ch.ID,
			Title:       pickLocale(ch.Title, zh),
			Description: pickLocale(ch.Description, zh),
			Ordinal:     ch.Ordinal,
			Completed:   isDone,
			Unlocked:    true,
		})
	}
	c.JSON(http.StatusOK, out)
}

// HandleGetTemplate — GET /api/chapters/:id/template
// Returns the user-facing skeleton code plus the list of TODO markers.
func HandleGetTemplate(c *gin.Context, db *sql.DB) {
	id := c.Param("id")
	ch, ok := chapters.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown chapter"})
		return
	}
	code, err := ch.TemplateCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	todos, err := ch.TemplateTodos(preferZh(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	apiTodos := make([]Todo, len(todos))
	for i, t := range todos {
		apiTodos[i] = Todo{Line: t.Line, Hint: t.Hint}
	}
	c.JSON(http.StatusOK, Template{Code: code, Todos: apiTodos})
}

// HandleGetHint — GET /api/chapters/:id/hint?line=N
// Returns the resolved (locale-specific) hint for the given 1-based line.
func HandleGetHint(c *gin.Context, db *sql.DB) {
	id := c.Param("id")
	ch, ok := chapters.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown chapter"})
		return
	}
	line, err := strconv.Atoi(c.Query("line"))
	if err != nil || line < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "line query param must be a positive integer"})
		return
	}
	h, found := ch.HintForLine(line)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "no hint for that line"})
		return
	}
	text := h.Hint.En
	if preferZh(c) {
		text = h.Hint.Zh
	}
	c.JSON(http.StatusOK, HintResponse{Text: text})
}

// HandleGetSolution — GET /api/chapters/:id/solution
// Returns the reference solution's source for on-demand viewing in the
// chapter detail view's answer modal.
func HandleGetSolution(c *gin.Context, db *sql.DB) {
	id := c.Param("id")
	ch, ok := chapters.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown chapter"})
		return
	}
	code, err := ch.SolutionCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, SolutionResponse{Code: code})
}

// HandleReset — POST /api/reset
// Wipes all progress so the user can start over. Returns 204 on success.
func HandleReset(c *gin.Context, db *sql.DB) {
	if _, err := db.Exec(`DELETE FROM progress`); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// pickLocale returns zh or en based on preferZh.
func pickLocale(l chapters.Locale, zh bool) string {
	if zh {
		return l.Zh
	}
	return l.En
}
