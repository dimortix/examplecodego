package planner

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func generateFloorPlanURL(seed int64, style string, rooms int, area int) string {
	return fmt.Sprintf("https://placehold.co/800x600/e2e8f0/1e293b?text=План+%s+%d+комнат+%dм²+%d",
		style, rooms, area, seed)
}

func generate3DRenderURL(seed int64, style string, rooms int, area int) string {
	return fmt.Sprintf("https://placehold.co/800x600/f8fafc/475569?text=3D+%s+%d+комнат+%dм²+%d",
		style, rooms, area, seed)
}

var styleTitles = map[string]string{
	"modern":       "Современный",
	"minimalist":   "Минималистичный",
	"scandinavian": "Скандинавский",
	"loft":         "Лофт",
	"classic":      "Классический",
	"provence":     "Прованс",
}

type Room struct {
	Name   string  `json:"name"`
	Area   float64 `json:"area"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

type FloorPlanGenerator struct {
	TotalArea float64
	Rooms     int
	Style     string
	Features  []string
}

func GeneratePlanHandler(c *gin.Context) {
	var req PlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	if req.Area < 20 || req.Area > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Площадь должна быть от 20 до 200 м²"})
		return
	}

	if req.Rooms < 1 || req.Rooms > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Количество комнат должно быть от 1 до 5"})
		return
	}

	aiPlanner := NewAIPlanner()

	log.Printf("Запрос на генерацию плана: %+v", req)

	if useAIMode := true; useAIMode {
		aiReq := AIGenerationRequest{
			Prompt:   fmt.Sprintf("Generate a %d square meter apartment with %d rooms in %s style", req.Area, req.Rooms, req.Style),
			Area:     req.Area,
			Rooms:    req.Rooms,
			Style:    req.Style,
			Features: req.Features,
		}

		aiResp, err := aiPlanner.GeneratePlan(aiReq)
		if err == nil && aiResp != nil {
			plans := generatePlansFromAI(aiResp, req)
			log.Printf("Успешно сгенерированы планы с использованием AI: %d планов", len(plans))
			c.JSON(http.StatusOK, plans)
			return
		}

		log.Printf("Не удалось использовать AI для генерации, ошибка: %v. Используем локальную генерацию.", err)
	}

	generator := &FloorPlanGenerator{
		TotalArea: float64(req.Area),
		Rooms:     req.Rooms,
		Style:     req.Style,
		Features:  req.Features,
	}

	plans := generator.GeneratePlans()
	log.Printf("Сгенерированы планы локально: %d планов", len(plans))

	c.JSON(http.StatusOK, plans)
}

func generatePlansFromAI(aiResp *AIGenerationResponse, req PlanRequest) []PlanResponse {
	now := time.Now().Format(time.RFC3339)

	styleTitle := styleTitles[req.Style]
	if styleTitle == "" {
		styleTitle = "Современный"
	}

	budgetRooms := modifyRoomSizes(aiResp.RoomData, 0.8)
	budgetRoomsJSON, _ := json.Marshal(budgetRooms)

	standardRoomsJSON, _ := json.Marshal(aiResp.RoomData)

	premiumRooms := modifyRoomSizes(aiResp.RoomData, 1.2)
	premiumRoomsJSON, _ := json.Marshal(premiumRooms)

	floorPlanURL := aiResp.FloorPlanURL
	render3DURL := aiResp.Render3DURL

	if floorPlanURL == "" {
		floorPlanURL = generateFloorPlanURL(rand.Int63(), req.Style, req.Rooms, req.Area)
	}

	if render3DURL == "" {
		render3DURL = generate3DRenderURL(rand.Int63(), req.Style, req.Rooms, req.Area)
	}

	plans := []PlanResponse{
		{
			ID:        uuid.New().String(),
			Title:     fmt.Sprintf("Бюджетный %s", styleTitle),
			Area:      req.Area,
			Rooms:     req.Rooms,
			Style:     req.Style,
			Features:  req.Features,
			FloorPlan: floorPlanURL,
			Render3D:  render3DURL,
			CreatedAt: now,
			UpdatedAt: now,
			RoomData:  string(budgetRoomsJSON),
		},
		{
			ID:        uuid.New().String(),
			Title:     fmt.Sprintf("Стандартный %s", styleTitle),
			Area:      req.Area,
			Rooms:     req.Rooms,
			Style:     req.Style,
			Features:  req.Features,
			FloorPlan: floorPlanURL,
			Render3D:  render3DURL,
			CreatedAt: now,
			UpdatedAt: now,
			RoomData:  string(standardRoomsJSON),
		},
		{
			ID:        uuid.New().String(),
			Title:     fmt.Sprintf("Премиум %s", styleTitle),
			Area:      req.Area,
			Rooms:     req.Rooms,
			Style:     req.Style,
			Features:  req.Features,
			FloorPlan: floorPlanURL,
			Render3D:  render3DURL,
			CreatedAt: now,
			UpdatedAt: now,
			RoomData:  string(premiumRoomsJSON),
		},
	}

	return plans
}

func modifyRoomSizes(rooms []Room, factor float64) []Room {
	modifiedRooms := make([]Room, len(rooms))

	for i, room := range rooms {
		modifiedRooms[i] = Room{
			Name:   room.Name,
			Area:   room.Area * factor,
			Width:  room.Width * factor,
			Height: room.Height * factor,
			X:      room.X,
			Y:      room.Y,
		}
	}

	return modifiedRooms
}

func (g *FloorPlanGenerator) GeneratePlans() []PlanResponse {
	rand.Seed(time.Now().UnixNano())
	now := time.Now().Format(time.RFC3339)
	styleTitle := styleTitles[g.Style]
	if styleTitle == "" {
		styleTitle = "Современный"
	}

	budgetRooms := g.generateRooms(0.8)
	standardRooms := g.generateRooms(1.0)
	premiumRooms := g.generateRooms(1.2)
	budgetRoomsJSON, _ := json.Marshal(budgetRooms)
	standardRoomsJSON, _ := json.Marshal(standardRooms)
	premiumRoomsJSON, _ := json.Marshal(premiumRooms)

	plans := []PlanResponse{
		{
			ID:        uuid.New().String(),
			Title:     fmt.Sprintf("Бюджетный %s", styleTitle),
			Area:      int(g.TotalArea),
			Rooms:     g.Rooms,
			Style:     g.Style,
			Features:  g.Features,
			FloorPlan: generateFloorPlanURL(rand.Int63(), g.Style, g.Rooms, int(g.TotalArea)),
			Render3D:  generate3DRenderURL(rand.Int63(), g.Style, g.Rooms, int(g.TotalArea)),
			CreatedAt: now,
			UpdatedAt: now,
			RoomData:  string(budgetRoomsJSON),
		},
		{
			ID:        uuid.New().String(),
			Title:     fmt.Sprintf("Стандартный %s", styleTitle),
			Area:      int(g.TotalArea),
			Rooms:     g.Rooms,
			Style:     g.Style,
			Features:  g.Features,
			FloorPlan: generateFloorPlanURL(rand.Int63(), g.Style, g.Rooms, int(g.TotalArea)),
			Render3D:  generate3DRenderURL(rand.Int63(), g.Style, g.Rooms, int(g.TotalArea)),
			CreatedAt: now,
			UpdatedAt: now,
			RoomData:  string(standardRoomsJSON),
		},
		{
			ID:        uuid.New().String(),
			Title:     fmt.Sprintf("Премиум %s", styleTitle),
			Area:      int(g.TotalArea),
			Rooms:     g.Rooms,
			Style:     g.Style,
			Features:  g.Features,
			FloorPlan: generateFloorPlanURL(rand.Int63(), g.Style, g.Rooms, int(g.TotalArea)),
			Render3D:  generate3DRenderURL(rand.Int63(), g.Style, g.Rooms, int(g.TotalArea)),
			CreatedAt: now,
			UpdatedAt: now,
			RoomData:  string(premiumRoomsJSON),
		},
	}

	return plans
}

func (g *FloorPlanGenerator) generateRooms(factor float64) []Room {
	rooms := make([]Room, 0)

	minRoomArea := 8.0

	remainingArea := g.TotalArea * factor

	bathroomArea := minRoomArea + rand.Float64()*4.0
	bathroom := Room{
		Name:   "Ванная",
		Area:   bathroomArea,
		Width:  2.0 + rand.Float64(),
		Height: bathroomArea / (2.0 + rand.Float64()),
		X:      0,
		Y:      0,
	}
	rooms = append(rooms, bathroom)
	remainingArea -= bathroomArea

	kitchenArea := minRoomArea + rand.Float64()*8.0
	kitchen := Room{
		Name:   "Кухня",
		Area:   kitchenArea,
		Width:  3.0 + rand.Float64()*2.0,
		Height: kitchenArea / (3.0 + rand.Float64()*2.0),
		X:      bathroom.Width,
		Y:      0,
	}
	rooms = append(rooms, kitchen)
	remainingArea -= kitchenArea

	hallwayArea := minRoomArea + rand.Float64()*4.0
	hallway := Room{
		Name:   "Коридор",
		Area:   hallwayArea,
		Width:  2.0 + rand.Float64(),
		Height: hallwayArea / (2.0 + rand.Float64()),
		X:      bathroom.Width + kitchen.Width,
		Y:      0,
	}
	rooms = append(rooms, hallway)
	remainingArea -= hallwayArea

	livingRoomsCount := g.Rooms - 2
	if livingRoomsCount <= 0 {
		livingRoomsCount = 1
	}

	areaPerRoom := remainingArea / float64(livingRoomsCount)

	livingRoomArea := areaPerRoom * 1.5
	livingRoom := Room{
		Name:   "Гостиная",
		Area:   livingRoomArea,
		Width:  4.0 + rand.Float64()*2.0,
		Height: livingRoomArea / (4.0 + rand.Float64()*2.0),
		X:      0,
		Y:      bathroom.Height + 1.0,
	}
	rooms = append(rooms, livingRoom)

	for i := 1; i < livingRoomsCount; i++ {
		bedroomArea := areaPerRoom
		bedroomName := fmt.Sprintf("Спальня %d", i)

		bedroom := Room{
			Name:   bedroomName,
			Area:   bedroomArea,
			Width:  3.5 + rand.Float64()*1.5,
			Height: bedroomArea / (3.5 + rand.Float64()*1.5),
			X:      livingRoom.Width * float64(i%2),
			Y:      bathroom.Height + livingRoom.Height + float64(i/2)*3.0,
		}
		rooms = append(rooms, bedroom)
	}

	return rooms
}

func SavePlanHandler(c *gin.Context) {
	var plan PlanResponse
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	plan.ID = uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	plan.CreatedAt = now
	plan.UpdatedAt = now

	c.JSON(http.StatusOK, plan)
}

func GetUserPlansHandler(c *gin.Context) {
	c.JSON(http.StatusOK, []PlanResponse{})
}
