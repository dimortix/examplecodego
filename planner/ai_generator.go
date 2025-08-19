package planner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type AIPlanner struct {
	APIToken    string
	APIEndpoint string
	HTTPClient  *http.Client
}

type AIGenerationRequest struct {
	Prompt   string   `json:"prompt"`
	Area     int      `json:"area"`
	Rooms    int      `json:"rooms"`
	Style    string   `json:"style"`
	Features []string `json:"features"`
}

type AIGenerationResponse struct {
	RoomData     []Room `json:"room_data"`
	FloorPlanURL string `json:"floor_plan_url"`
	Render3DURL  string `json:"render_3d_url"`
	Error        string `json:"error,omitempty"`
}

func NewAIPlanner() *AIPlanner {
	apiToken := os.Getenv("HUGGINGFACE_API_TOKEN")

	if apiToken == "" {
		apiToken = "тут писал токен в коде,но с появлением render больше так не делаю,старый код"
		log.Println("HUGGINGFACE_API_TOKEN не найден, используем резервный токен")
	}

	if apiToken == "" {
		log.Println("HUGGINGFACE_API_TOKEN не найден, используем тестовый режим")
	}

	return &AIPlanner{
		APIToken:    apiToken,
		APIEndpoint: "https://api-inference.huggingface.co/models/",
		HTTPClient: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (ap *AIPlanner) GeneratePlan(req AIGenerationRequest) (*AIGenerationResponse, error) {
	if ap.APIToken == "" {
		return ap.mockGeneratePlan(req)
	}
	prompt := fmt.Sprintf("Generate a detailed apartment floor plan with %d square meters, %d rooms, in %s style",
		req.Area, req.Rooms, req.Style)

	if len(req.Features) > 0 {
		prompt += ". Features: " + strings.Join(req.Features, ", ")
	}
	requestBody, err := json.Marshal(map[string]interface{}{
		"inputs": prompt,
		"parameters": map[string]interface{}{
			"negative_prompt":     "poor quality, low resolution, bad architecture, unrealistic layout",
			"num_inference_steps": 30,
			"guidance_scale":      7.5,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка при формировании запроса: %v", err)
	}

	modelEndpoint := "black-forest-labs/FLUX.1-dev"

	log.Printf("Используем модель Hugging Face: %s", modelEndpoint)

	request, err := http.NewRequest("POST", ap.APIEndpoint+modelEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании HTTP запроса: %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+ap.APIToken)
	request.Header.Set("Content-Type", "application/json")

	log.Printf("Отправляем запрос к %s с заголовком Authorization", ap.APIEndpoint+modelEndpoint)
	resp, err := ap.HTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении HTTP запроса: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Получен ответ от API, статус: %d", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка API: %s, статус: %d", string(body), resp.StatusCode)
	}
	log.Printf("Успешно получен ответ от модели Hugging Face")

	contentType := resp.Header.Get("Content-Type")
	log.Printf("Тип контента ответа: %s", contentType)

	timestamp := time.Now().Unix()
	floorPlanFilename := fmt.Sprintf("floorplan_%d_%d_%d.png", req.Area, req.Rooms, timestamp)
	render3DFilename := fmt.Sprintf("render3d_%d_%d_%d.png", req.Area, req.Rooms, timestamp)

	staticBasePath := "./static/plans/"
	os.MkdirAll(staticBasePath, os.ModePerm)
	floorPlanPath := staticBasePath + floorPlanFilename
	_ = staticBasePath + render3DFilename
	floorPlanURL := "/static/plans/" + floorPlanFilename
	render3DURL := "/static/plans/" + render3DFilename

	err = ioutil.WriteFile(floorPlanPath, body, 0644)
	if err != nil {
		log.Printf("Ошибка при сохранении изображения плана: %v", err)
	}

	generator := &FloorPlanGenerator{
		TotalArea: float64(req.Area),
		Rooms:     req.Rooms,
		Style:     req.Style,
		Features:  req.Features,
	}

	rooms := generator.generateRooms(1.0)

	return &AIGenerationResponse{
		RoomData:     rooms,
		FloorPlanURL: floorPlanURL,
		Render3DURL:  render3DURL,
	}, nil
}

func (ap *AIPlanner) mockGeneratePlan(req AIGenerationRequest) (*AIGenerationResponse, error) {
	log.Printf("Используем мок-генерацию плана для: %+v", req)

	generator := &FloorPlanGenerator{
		TotalArea: float64(req.Area),
		Rooms:     req.Rooms,
		Style:     req.Style,
		Features:  req.Features,
	}

	rooms := generator.generateRooms(1.0)

	seed := time.Now().UnixNano()
	floorPlanURL := generateFloorPlanURL(seed, req.Style, req.Rooms, req.Area)
	render3DURL := generate3DRenderURL(seed, req.Style, req.Rooms, req.Area)

	return &AIGenerationResponse{
		RoomData:     rooms,
		FloorPlanURL: floorPlanURL,
		Render3DURL:  render3DURL,
	}, nil
}

func (ap *AIPlanner) GenerateInteriorDesign(roomType, style string) (string, error) {
	if ap.APIToken == "" {
		return fmt.Sprintf("https://source.unsplash.com/random/1200x800/?%s,%s,interior", roomType, style), nil
	}
	prompt := fmt.Sprintf("Realistic interior design of a %s in %s style, professional photography, detailed, natural lighting, 8k resolution", roomType, style)

	requestBody, err := json.Marshal(map[string]interface{}{
		"inputs": prompt,
		"parameters": map[string]interface{}{
			"negative_prompt":     "low quality, blurry, distorted, disfigured, poor design, unrealistic, cartoon",
			"num_inference_steps": 40,
			"guidance_scale":      8.0,
		},
	})
	if err != nil {
		return "", err
	}

	modelEndpoint := "black-forest-labs/FLUX.1-dev"

	request, err := http.NewRequest("POST", ap.APIEndpoint+modelEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	request.Header.Set("Authorization", "Bearer "+ap.APIToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := ap.HTTPClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		bodyText, _ := ioutil.ReadAll(response.Body)
		return "", fmt.Errorf("API error: %s, status: %d", string(bodyText), response.StatusCode)
	}
	timestamp := time.Now().Unix()
	interiorFilename := fmt.Sprintf("interior_%s_%s_%d.png", roomType, style, timestamp)

	staticBasePath := "./static/interiors/"
	os.MkdirAll(staticBasePath, os.ModePerm)
	interiorPath := staticBasePath + interiorFilename

	interiorURL := "/static/interiors/" + interiorFilename

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	err = ioutil.WriteFile(interiorPath, body, 0644)
	if err != nil {
		log.Printf("Ошибка при сохранении изображения интерьера: %v", err)
		return fmt.Sprintf("https://source.unsplash.com/random/1200x800/?%s,%s,interior", roomType, style), nil
	}

	return interiorURL, nil
}
