package entity

type Server struct {
	ID       int      `json:"id"`
	Label    string   `json:"label"`
	IP       string   `json:"ip"`
	RootUser string   `json:"root_user"`
	Password string   `json:"password"`
	Purpose  string   `json:"purpose"`
	Tags     []string `json:"tags"`
}

func (s Server) GetID() int {
	return s.ID
}

func (s Server) SetID(id int) Server {
	s.ID = id
	return s
}
