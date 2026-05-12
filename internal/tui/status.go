package tui

type StatusModel struct {
	message    string
	processing bool
}

func NewStatusModel() StatusModel {
	return StatusModel{}
}

func (m StatusModel) View(processing bool) string {
	return renderStatus(processing)
}

func renderStatus(processing bool) string {
	if processing {
		return "⟳ Processing..."
	}
	return "─"
}
