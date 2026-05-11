package styles

import "github.com/charmbracelet/lipgloss"

// SequoiaTree returns a colored pixel-art sequoia tree string using full-block
// (▄, █) characters. The canopy rows use forest green (#228B22) and the trunk
// and root flare rows use saddle brown (#8B4513).
func SequoiaTree() string {
	canopy := lipgloss.NewStyle().Foreground(colorFoliage)
	bark := lipgloss.NewStyle().Foreground(colorBark)

	return canopy.Render("           ▄██▄") + "\n" +
		canopy.Render("         ▄██████▄") + "\n" +
		canopy.Render("        ▄████████▄") + "\n" +
		canopy.Render("       ▄██████████▄") + "\n" +
		canopy.Render("     ▄████████████▄") + "\n" +
		canopy.Render("    ▄██████████████▄") + "\n" +
		canopy.Render("   ▄████████████████▄") + "\n" +
		bark.Render("          ██████") + "\n" +
		bark.Render("          ██████") + "\n" +
		bark.Render("          ██████") + "\n" +
		bark.Render("          ██████") + "\n" +
		bark.Render("          ██████") + "\n" +
		bark.Render("         ▄██████▄") + "\n" +
		bark.Render("        ▄████████▄") + "\n"
}
