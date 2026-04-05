package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	Port          string
	BotToken      string
	ServerTimeout time.Duration
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	return &Config{
		Port:          port,
		BotToken:      "8271246541:AAEi22YopJfQiKBvXgGv3y4wcyqnRoyquYw",
		ServerTimeout: 30 * time.Second,
	}
}

type AnalysisResult struct {
	Name         string   `json:"name"`
	Age          int      `json:"age"`
	Education    string   `json:"education"`
	Experience   string   `json:"experience"`
	Motivation   string   `json:"motivation"`
	Skills       []string `json:"skills"`
	SoftSkills   []string `json:"soft_skills"`
	Achievements []string `json:"achievements"`
	Pros         []string `json:"pros"`
	Cons         []string `json:"cons"`
	Questions    []string `json:"questions"`
	Score        int      `json:"score"`
	Recommend    string   `json:"recommend"`
}

type TelegramUpdate struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

func analyzeEssay(essay string) *AnalysisResult {
	text := strings.ToLower(essay)

	name := extractName(essay)
	age := extractAge(essay)
	education := extractEducation(text)
	experience := extractExperience(text)
	motivation := extractMotivation(text)
	skills := extractSkills(text)
	softSkills := extractSoftSkills(text)
	achievements := extractAchievements(text)

	score := calculateScore(skills, achievements, motivation)

	return &AnalysisResult{
		Name:         name,
		Age:          age,
		Education:    education,
		Experience:   experience,
		Motivation:   motivation,
		Skills:       skills,
		SoftSkills:   softSkills,
		Achievements: achievements,
		Pros:         generatePros(skills, achievements, motivation),
		Cons:         generateCons(education, experience),
		Questions:    generateQuestions(skills, achievements),
		Score:        score,
		Recommend:    generateRecommendation(score),
	}
}

func calculateScore(skills []string, achievements []string, motivation string) int {
	score := 50

	if len(skills) >= 3 {
		score += 20
	} else if len(skills) >= 1 {
		score += 10
	}

	if len(achievements) >= 2 {
		score += 20
	} else if len(achievements) >= 1 {
		score += 10
	}

	if strings.Contains(motivation, "высокая") {
		score += 10
	}

	if score > 100 {
		score = 100
	}
	return score
}

func generatePros(skills []string, achievements []string, motivation string) []string {
	pros := []string{}

	if len(skills) >= 2 {
		pros = append(pros, "Хороший набор технических навыков")
	}
	if len(achievements) >= 1 {
		pros = append(pros, "Есть конкретные достижения")
	}
	if strings.Contains(motivation, "высокая") {
		pros = append(pros, "Высокая мотивация")
	}
	if len(pros) == 0 {
		pros = append(pros, "Есть потенциал для развития")
	}
	return pros
}

func generateCons(education string, experience string) []string {
	cons := []string{}

	if education == "не указано" {
		cons = append(cons, "Не указано образование")
	}
	if experience == "не указан" {
		cons = append(cons, "Не указан опыт")
	}
	if len(cons) == 0 {
		cons = append(cons, "Требуется собеседование")
	}
	return cons
}

func generateQuestions(skills []string, achievements []string) []string {
	questions := []string{
		"Почему вы выбрали нашу программу?",
		"Какие ваши карьерные цели?",
	}

	if len(skills) > 0 {
		questions = append(questions, "Расскажите о ваших навыках подробнее")
	}
	if len(achievements) > 0 {
		questions = append(questions, "Какое достижение считаете самым значимым?")
	}
	return questions
}

func generateRecommendation(score int) string {
	if score >= 80 {
		return "Сильный кандидат, рекомендуется к зачислению"
	} else if score >= 65 {
		return "Хороший кандидат, рекомендуется к рассмотрению"
	} else {
		return "Требуется дополнительное собеседование"
	}
}

type BotService struct {
	config *Config
	apiURL string
	client *http.Client
}

func NewBotService(cfg *Config) *BotService {
	return &BotService{
		config: cfg,
		apiURL: fmt.Sprintf("https://api.telegram.org/bot%s", cfg.BotToken),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (b *BotService) Start() {
	log.Println("[Bot] Telegram бот запущен")
	offset := 0

	for {
		updates, err := b.getUpdates(offset)
		if err != nil {
			log.Printf("[Bot] Ошибка: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range updates {
			b.handleUpdate(update)
			offset = update.UpdateID + 1
		}
	}
}

func (b *BotService) getUpdates(offset int) ([]TelegramUpdate, error) {
	url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=30", b.apiURL, offset)
	resp, err := b.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result []TelegramUpdate `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Result, nil
}

func (b *BotService) handleUpdate(update TelegramUpdate) {
	if update.Message.Text == "" {
		return
	}

	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	if text == "/start" {
		b.sendMessage(chatID, b.getStartMessage())
		return
	}

	b.sendChatAction(chatID)

	result := analyzeEssay(text)
	b.sendMessage(chatID, b.formatResponse(result))
}

func (b *BotService) getStartMessage() string {
	return " *inVision U AI Бот*\n\n" +
		"Отправьте текст мотивационного письма, и я проанализирую его.\n\n" +
		"*Что я умею:*\n" +
		" Анализ мотивации\n" +
		" Оценка навыков\n" +
		" Выявление достижений\n" +
		" Рекомендация"
}

func (b *BotService) formatResponse(result *AnalysisResult) string {
	return fmt.Sprintf(
		" *Результат анализа*\n\n"+
			" *ФИО:* %s\n"+
			" *Возраст:* %d\n"+
			" *Образование:* %s\n"+
			" *Опыт:* %s\n"+
			" *Мотивация:* %s\n\n"+
			" *Навыки:* %s\n"+
			" *Soft skills:* %s\n"+
			" *Достижения:* %s\n\n"+
			" *Плюсы:* %s\n"+
			"️ *Минусы:* %s\n\n"+
			" *Вопросы:*\n%s\n\n"+
			" *Скор:* %d/100\n\n"+
			" *Рекомендация:* %s",
		result.Name, result.Age, result.Education, result.Experience, result.Motivation,
		joinStrings(result.Skills), joinStrings(result.SoftSkills), joinStrings(result.Achievements),
		joinStrings(result.Pros), joinStrings(result.Cons),
		joinStrings(result.Questions),
		result.Score, result.Recommend,
	)
}

func (b *BotService) sendMessage(chatID int64, text string) {
	url := fmt.Sprintf("%s/sendMessage", b.apiURL)
	body := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	jsonData, _ := json.Marshal(body)
	b.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
}

func (b *BotService) sendChatAction(chatID int64) {
	url := fmt.Sprintf("%s/sendChatAction", b.apiURL)
	body := map[string]interface{}{
		"chat_id": chatID,
		"action":  "typing",
	}
	jsonData, _ := json.Marshal(body)
	b.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
}

type Server struct {
	config     *Config
	httpServer *http.Server
	botService *BotService
}

func NewServer(cfg *Config) *Server {
	return &Server{
		config:     cfg,
		botService: NewBotService(cfg),
	}
}

func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      s.setupRoutes(),
		ReadTimeout:  s.config.ServerTimeout,
		WriteTimeout: s.config.ServerTimeout,
		IdleTimeout:  60 * time.Second,
	}

	go s.botService.Start()

	go func() {
		log.Printf("Сервер запущен на http://localhost:%s", s.config.Port)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP сервер упал: %v", err)
		}
	}()

	return nil
}

func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/analyze", s.handleAnalyze)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	return s.corsMiddleware(s.loggingMiddleware(mux))
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	var req struct {
		Essay string `json:"essay"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Неверный JSON")
		return
	}

	result := analyzeEssay(req.Essay)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func joinStrings(arr []string) string {
	if len(arr) == 0 {
		return "—"
	}
	return strings.Join(arr, ", ")
}

func extractName(essay string) string {
	lines := strings.Split(essay, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && len(line) < 50 {
			for _, ch := range line {
				if (ch >= 'А' && ch <= 'Я') || (ch >= 'а' && ch <= 'я') {
					return line
				}
			}
		}
	}
	return "Кандидат"
}

func extractAge(essay string) int {
	words := strings.Fields(essay)
	for _, w := range words {
		if age, err := strconv.Atoi(w); err == nil && age >= 16 && age <= 80 {
			return age
		}
	}
	return 0
}

func extractEducation(text string) string {
	if strings.Contains(text, "университет") || strings.Contains(text, "вуз") || strings.Contains(text, "высшее") {
		return "Высшее образование"
	}
	if strings.Contains(text, "школ") || strings.Contains(text, "класс") {
		return "Среднее образование"
	}
	return "не указано"
}

func extractExperience(text string) string {
	if strings.Contains(text, "курсы") || strings.Contains(text, "стажировка") {
		return "Есть дополнительное обучение"
	}
	if strings.Contains(text, "проект") || strings.Contains(text, "хакатон") {
		return "Есть опыт проектной деятельности"
	}
	return "не указан"
}

func extractMotivation(text string) string {
	if strings.Contains(text, "хочу") && strings.Contains(text, "развитие") {
		return "Высокая мотивация, четкое понимание целей"
	}
	if strings.Contains(text, "интерес") || strings.Contains(text, "нравится") {
		return "Хорошая мотивация"
	}
	return "Мотивация прослеживается"
}

func extractSkills(text string) []string {
	skills := []string{}
	skillList := []string{"go", "python", "java", "javascript", "flutter", "react", "sql", "git", "docker"}
	for _, s := range skillList {
		if strings.Contains(text, s) {
			skills = append(skills, strings.ToUpper(s[:1])+s[1:])
		}
	}
	if len(skills) == 0 {
		skills = append(skills, "Базовые навыки программирования")
	}
	return skills
}

func extractSoftSkills(text string) []string {
	soft := []string{}
	if strings.Contains(text, "команд") {
		soft = append(soft, "Работа в команде")
	}
	if strings.Contains(text, "ответствен") {
		soft = append(soft, "Ответственность")
	}
	if strings.Contains(text, "обучаем") {
		soft = append(soft, "Обучаемость")
	}
	if strings.Contains(text, "аналитическ") {
		soft = append(soft, "Аналитическое мышление")
	}
	return soft
}

func extractAchievements(text string) []string {
	ach := []string{}
	if strings.Contains(text, "хакатон") {
		ach = append(ach, "Участие в хакатоне")
	}
	if strings.Contains(text, "проект") {
		ach = append(ach, "Реализованные проекты")
	}
	if strings.Contains(text, "олимпиад") {
		ach = append(ach, "Участие в олимпиадах")
	}
	return ach
}

func main() {
	cfg := LoadConfig()
	server := NewServer(cfg)

	if err := server.Start(); err != nil {
		log.Fatalf("Ошибка запуска: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Завершение работы...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при завершении: %v", err)
	}

	log.Println("Сервер остановлен")
}

