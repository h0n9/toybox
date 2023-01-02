package client

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/spf13/cobra"
)

type boxL struct {
	views.BoxLayout
}

var (
	app = &views.Application{}
	box = &boxL{}
)

func (m *boxL) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape:
			fallthrough
		case tcell.KeyCtrlC:
			app.Quit()
			return true
		}
	}
	return false
}

var Cmd = &cobra.Command{
	Use:   "client",
	Short: "run msg lake client (lab)",
	RunE: func(cmd *cobra.Command, args []string) error {
		title := views.NewTextBar()
		title.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow))
		title.SetCenter("msg lake (prototype)", tcell.StyleDefault)

		chatBox := views.NewText()
		chatBox.SetText("chat box")
		chatBox.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorLime))
		chatBox.SetAlignment(views.VAlignCenter | views.HAlignCenter)

		inputBox := views.NewText()
		inputBox.SetText("input box")
		inputBox.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue))
		inputBox.SetAlignment(views.VAlignCenter | views.HAlignLeft)

		box.SetOrientation(views.Vertical)
		box.AddWidget(title, 0)
		box.AddWidget(chatBox, 0.5)
		box.AddWidget(inputBox, 0)

		app.SetRootWidget(box)
		err := app.Run()
		if err != nil {
			return err
		}

		return nil
	},
}
