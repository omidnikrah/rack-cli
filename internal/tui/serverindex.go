package tui

import "github.com/omidnikrah/rack-cli/internal/entity"

const deletedServerLabel = "deleted server"

type serverIndex struct {
	labels map[int]string
}

func (x *serverIndex) set(rows []entity.Server) {
	x.labels = make(map[int]string, len(rows))
	for _, s := range rows {
		x.labels[s.ID] = hostLabel(s)
	}
}

func (x *serverIndex) label(id int) string {
	if label, ok := x.labels[id]; ok {
		return label
	}

	return deletedServerLabel
}

func (x *serverIndex) exists(id int) bool {
	_, ok := x.labels[id]
	return ok
}
