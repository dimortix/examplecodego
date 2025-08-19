package planner

type PlanRequest struct {
	Area     int      `json:"area"`
	Rooms    int      `json:"rooms"`
	Style    string   `json:"style"`
	Features []string `json:"features"`
}

type PlanResponse struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Area      int      `json:"area"`
	Rooms     int      `json:"rooms"`
	Style     string   `json:"style"`
	Features  []string `json:"features"`
	FloorPlan string   `json:"floor_plan"`
	Render3D  string   `json:"render_3d"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	RoomData  string   `json:"room_data"`
}
