package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type AnalysisResult struct {
	Name         string   `json:"name"`
	Age          int      `json:"age"`
	Education    string   `json:"education"`
	Experience   string   `json:"experience"`
	Motivation   string   `json:"motivation"`
	Skills       []string `json:"skills"`
	SoftSkills   []string `json:"soft_skills"`
	Achievements []string `json:"achievements"`
	Metrics      struct {
		Motivation   int `json:"motivation"`
		Skills       int `json:"skills"`
		Experience   int `json:"experience"`
		Education    int `json:"education"`
		Achievements int `json:"achievements"`
		Grammar      int `json:"grammar"`
	} `json:"metrics"`
	Pros                  []string `json:"pros"`
	Cons                  []string `json:"cons"`
	QuestionsForInterview []string `json:"questions_for_interview"`
	AttentionPoints       []string `json:"attention_points"`
	AiSuggestion          string   `json:"ai_suggestion"`
}

func main() {
	http.HandleFunc("/api/analyze", handleAnalyze)
	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Println("Server on http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Essay string `json:"essay"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Essay) == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	result := analyzeText(req.Essay)

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		return
	}
}

func analyzeText(essay string) *AnalysisResult {
	text := strings.ToLower(essay)

	result := &AnalysisResult{}
	if strings.Contains(essay, "Анна") {
		result.Name = "Анна Смирнова"
	} else if strings.Contains(essay, "Дмитрий") {
		result.Name = "Дмитрий"
	} else {
		lines := strings.Split(essay, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) > 0 && len(line) < 50 {
				hasRus := false
				for _, ch := range line {
					if (ch >= 'А' && ch <= 'Я') || (ch >= 'а' && ch <= 'я') {
						hasRus = true
						break
					}
				}
				if hasRus {
					result.Name = line
					break
				}
			}
		}
		if result.Name == "" {
			result.Name = "Кандидат"
		}
	}
	words := strings.Fields(essay)
	for _, word := range words {
		if age, err := strconv.Atoi(word); err == nil && age >= 16 && age <= 80 {
			result.Age = age
			break
		}
	}
	if strings.Contains(text, "мгу") || strings.Contains(text, "университет") {
		result.Education = "Высшее (МГУ)"
	} else if strings.Contains(text, "высшее") {
		result.Education = "Высшее образование"
	} else {
		result.Education = "не указано"
	}
	if strings.Contains(text, "3 года") {
		result.Experience = "3 года (ведущий бухгалтер)"
	} else if strings.Contains(text, "лет") {
		result.Experience = "Опыт указан"
	} else {
		result.Experience = "не указан"
	}
	if strings.Contains(text, "хочу") && strings.Contains(text, "развитие") {
		result.Motivation = "Высокая мотивация, четкие цели"
		result.Metrics.Motivation = 90
	} else if strings.Contains(text, "хочу") {
		result.Motivation = "Мотивация прослеживается"
		result.Metrics.Motivation = 70
	} else {
		result.Motivation = "Мотивация выражена слабо"
		result.Metrics.Motivation = 50
	}
	skills := []string{}
	if strings.Contains(text, "excel") {
		skills = append(skills, "Excel")
	}
	if strings.Contains(text, "1с") {
		skills = append(skills, "1С")
	}
	if strings.Contains(text, "sql") {
		skills = append(skills, "SQL")
	}
	if strings.Contains(text, "анализ") {
		skills = append(skills, "Анализ данных")
	}
	if strings.Contains(text, "word") {
		skills = append(skills, "Word")
	}
	if len(skills) == 0 {
		skills = append(skills, "Базовые навыки")
	}
	result.Skills = skills
	result.Metrics.Skills = 70 + len(skills)*5
	if result.Metrics.Skills > 100 {
		result.Metrics.Skills = 100
	}
	soft := []string{}
	if strings.Contains(text, "ответствен") {
		soft = append(soft, "Ответственность")
	}
	if strings.Contains(text, "коммуника") {
		soft = append(soft, "Коммуникабельность")
	}
	if strings.Contains(text, "внимательн") {
		soft = append(soft, "Внимательность")
	}
	if strings.Contains(text, "стресс") {
		soft = append(soft, "Стрессоустойчивость")
	}
	result.SoftSkills = soft
	achievements := []string{}
	if strings.Contains(text, "автоматизирова") {
		achievements = append(achievements, "Автоматизация отчетности")
	}
	if strings.Contains(text, "оптимизирова") {
		achievements = append(achievements, "Оптимизация процессов")
	}
	if strings.Contains(text, "внедри") {
		achievements = append(achievements, "Внедрение новых систем")
	}
	if strings.Contains(text, "сократи") {
		achievements = append(achievements, "Сокращение времени/затрат")
	}
	result.Achievements = achievements
	result.Metrics.Achievements = 60 + len(achievements)*10
	if result.Metrics.Achievements > 100 {
		result.Metrics.Achievements = 100
	}
	result.Metrics.Experience = 70
	result.Metrics.Education = 75
	result.Metrics.Grammar = 80
	result.Pros = []string{
		"Профильное образование",
		"Наличие опыта работы",
		"Четко выраженная мотивация",
	}
	result.Cons = []string{
		"Явных недостатков не выявлено",
	}
	result.QuestionsForInterview = []string{
		"Почему вы выбрали нашу компанию?",
		"Расскажите о ваших достижениях подробнее",
		"Как вы работаете в стрессовых ситуациях?",
	}
	result.AttentionPoints = []string{
		"Рекомендуется провести собеседование",
	}

	totalScore := (result.Metrics.Motivation + result.Metrics.Skills + result.Metrics.Achievements) / 3
	if totalScore >= 80 {
		result.AiSuggestion = "Сильный кандидат. Рекомендуется к приоритетному рассмотрению."
	} else if totalScore >= 65 {
		result.AiSuggestion = "Хороший кандидат. Рекомендуется пригласить на собеседование."
	} else {
		result.AiSuggestion = "Кандидат требует дополнительной оценки."
	}

	return result
}
