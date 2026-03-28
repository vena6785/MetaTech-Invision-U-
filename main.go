package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type AnalysisResult struct {
	Name         string   `json:"name"`
	Age          int      `json:"age"`
	Motivation   string   `json:"motivation"`
	Achievements []string `json:"achievements"`
	Skills       []string `json:"skills"`
	Score        int      `json:"score"`
	Strengths    []string `json:"strengths"`
	Weaknesses   []string `json:"weaknesses"`
	Recommend    string   `json:"recommend"`
}

func main() {
	http.HandleFunc("/api/analyze", withCORS(analyzeHandler))
	http.Handle("/", http.FileServer(http.Dir("./web")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Сервер запущен на http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}
func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}
	var req struct {
		Essay string `json:"essay"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}
	if strings.TrimSpace(req.Essay) == "" {
		writeError(w, http.StatusBadRequest, "Текст не может быть пустым")
		return
	}
	result, err := callAI(req.Essay)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Ошибка при обращении к ИИ: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		return
	}
}

func callAI(essay string) (*AnalysisResult, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return metaAI(essay), nil
	}

	prompt := `Проанализируй текст кандидата. Верни ТОЛЬКО JSON:
{
  "name": "имя",
  "age": возраст,
  "motivation": "мотивация",
  "achievements": ["достижение1", "достижение2"],
  "skills": ["навык1", "навык2"],
  "score": число от 0 до 100,
  "strengths": ["сильная сторона1", "сильная сторона2"],
  "weaknesses": ["слабая сторона1", "слабая сторона2"],
  "recommend": "рекомендация для комиссии"
}

Текст: ` + essay

	reqBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
	}

	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, nil
	}

	var result AnalysisResult
	if err := json.Unmarshal([]byte(openAIResp.Choices[0].Message.Content), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Пока что metaAI-это лишь тестовый режим, в будущем настроим полностью
func metaAI(essay string) *AnalysisResult {
	words := strings.Fields(essay)

	name := "не указано"
	if len(words) > 0 {
		name = words[0]
	}

	age := 0
	for _, w := range words {
		if a, err := strconv.Atoi(w); err == nil && a >= 10 && a <= 100 {
			age = a
			break
		}
	}

	return &AnalysisResult{
		Name:         name,
		Age:          age,
		Motivation:   "указана (тестовый режим)",
		Achievements: []string{"указаны (тестовый режим)"},
		Skills:       []string{"указаны (тестовый режим)"},
		Score:        70,
		Strengths:    []string{"требуется реальный AI для анализа"},
		Weaknesses:   []string{"требуется реальный AI для анализа"},
		Recommend:    "требуется подключение OpenAI API",
	}
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(map[string]string{
		"error":   message,
		"code":    strconv.Itoa(code),
		"success": "false",
	})
	if err != nil {
		return
	}
}
