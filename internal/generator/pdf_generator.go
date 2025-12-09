package generator

import (
	"fmt"
	"log"
	"math"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	marotoimg "github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/johnfercher/maroto/v2/pkg/repository"
)

// ============================================================================
// PAGE LAYOUT
// ============================================================================
const (
	pageWidth          = 612.0                                // 8.5 inches in points
	leftMargin         = 10.0
	rightMargin        = 10.0
	usableWidth        = pageWidth - leftMargin - rightMargin // 592pt
	avgCharWidthAt10pt = 5.5                                  // For line calculation
)

// ============================================================================
// ICONS & SEPARATORS
// ============================================================================
const (
	iconPercent        = 80.0 // Icon size percentage
	iconLeftOffset     = 0.0  // Left position of icons
	separatorThickness = 0.25 // Separator line thickness
)

// ============================================================================
// FONT SIZES
// ============================================================================
const (
	fontSizeName          = 20.0 // Name at top
	fontSizeTitle         = 14.0 // Professional title
	fontSizeSectionHeader = 12.0 // Section headers (SKILLS, EXPERIENCE)
	fontSizeJobTitle      = 11.0 // Job/project titles
	fontSizeSummary       = 10.0 // Summary paragraph
	fontSizeContact       = 10.0 // Contact info
	fontSizeSkills        = 10.0 // Skills text
	fontSizeBullet        = 9.0  // Achievement bullets
	fontSizeTech          = 9.0  // Tech stack
)

// ============================================================================
// ROW HEIGHTS - Fixed
// ============================================================================
const (
	rowHeightName          = 10.0 // Name row
	rowHeightSectionHeader = 9.0  // Section header rows
	rowHeightJobTitle      = 6.5  // Job/project title
	rowHeightTechStack     = 3.5  // Tech stack
)

// ============================================================================
// ROW HEIGHTS - Dynamic (base + perLine * numLines)
// ============================================================================
const (
	skillRowBase    = 2.0 // Skills base height
	skillRowPerLine = 2.5 // Skills per line

	bulletRowBase    = 1.8 // Bullet base height
	bulletRowPerLine = 2.4 // Bullet per line
)

// ============================================================================
// VERTICAL SPACING - Section Gaps
// ============================================================================
const (
	spaceAfterSummary = 3.0 // Gap before first section
)

// ============================================================================
// PADDING - Internal spacing within rows
// ============================================================================
const (
	sectionHeaderTopPadding = 2.5  // Top padding in section headers
	sectionHeaderLeftPadding = 10.0 // Left padding (after icon)
	nameTopPadding = 3.0
	titleTopPadding = 3.0
	summaryTopPadding = 2.0
	contactTopPadding = 1.0
)

// calculateLineCount estimates how many lines text will take based on width and font size
func calculateLineCount(text string, widthCols int, fontSize float64) int {
	// Calculate available width in points
	availableWidth := (float64(widthCols) / 12.0) * usableWidth

	// Estimate character width for this font size
	charWidth := avgCharWidthAt10pt * (fontSize / 10.0)

	// Calculate chars per line
	charsPerLine := int(availableWidth / charWidth)

	if charsPerLine == 0 {
		return 1
	}

	// Calculate number of lines needed
	textLen := len(text)
	lines := int(math.Ceil(float64(textLen) / float64(charsPerLine)))

	if lines == 0 {
		return 1
	}

	return lines
}

// GeneratePDF creates a PDF from a Resume struct
func GeneratePDF(resume *Resume, outputPath string) error {
	log.Printf("Generating PDF at %s...", outputPath)

	mrt := getMaroto(resume)
	document, err := mrt.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate PDF document: %w", err)
	}

	err = document.Save(outputPath)
	if err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	// TODO: Check page count and apply scaling if needed (Phase 6)
	// For now, just generate with default sizes

	log.Printf("PDF generated successfully")
	return nil
}

func getMaroto(resume *Resume) core.Maroto {
	// Load custom fonts
	customFonts, err := repository.New().
		AddUTF8Font("dejavu", fontstyle.Normal, "fonts/DejaVuSans.ttf").
		AddUTF8Font("dejavu", fontstyle.Bold, "fonts/DejaVuSans-Bold.ttf").
		AddUTF8Font("dejavu", fontstyle.Italic, "fonts/DejaVuSans-Oblique.ttf").
		AddUTF8Font("dejavu", fontstyle.BoldItalic, "fonts/DejaVuSans-BoldOblique.ttf").
		AddUTF8Font("dejavu-light", fontstyle.Normal, "fonts/DejaVuSans-ExtraLight.ttf").
		Load()
	if err != nil {
		log.Printf("Warning: Failed to load custom fonts: %v", err)
		// Fallback to default config
		cfg := config.NewBuilder().
			WithPageNumber().
			WithLeftMargin(10).
			WithTopMargin(10).
			WithRightMargin(10).
			Build()
		return maroto.New(cfg)
	}

	cfg := config.NewBuilder().
		WithCustomFonts(customFonts).
		WithDefaultFont(&props.Font{Family: "dejavu"}).
		WithPageNumber().
		WithLeftMargin(10).
		WithTopMargin(10).
		WithRightMargin(10).
		Build()

	mrt := maroto.New(cfg)

	// Header with name and title
	mrt.AddRows(text.NewRow(15, resume.Name, props.Text{
		Top:   3,
		Style: fontstyle.Bold,
		Size:  20,
		Align: align.Left,
	}))

	// Contact info in 2x3 grid with icons
	mrt.AddRow(8,
		marotoimg.NewFromFileCol(1, "icons-png/map-pin.png", props.Rect{
			Center:  false,
			Percent: 60,
			Top:     1,
			Left:    10,
		}),
		text.NewCol(3, resume.Contact.Location, props.Text{
			Size:  10,
			Align: align.Left,
			Top:   1,
		}),
		marotoimg.NewFromFileCol(1, "icons-png/phone.png", props.Rect{
			Center:  false,
			Percent: 60,
			Top:     1,
			Left:    10,
		}),
		text.NewCol(3, resume.Contact.Phone, props.Text{
			Size:  10,
			Align: align.Left,
			Top:   1,
		}),
		marotoimg.NewFromFileCol(1, "icons-png/mail.png", props.Rect{
			Center:  false,
			Percent: 60,
			Top:     1,
			Left:    10,
		}),
		text.NewCol(3, resume.Contact.Email, props.Text{
			Size:  10,
			Align: align.Left,
			Top:   1,
		}),
	)
	mrt.AddRow(8,
		marotoimg.NewFromFileCol(1, "icons-png/link.png", props.Rect{
			Center:  false,
			Percent: 60,
			Top:     1,
			Left:    10,
		}),
		text.NewCol(3, resume.Contact.Website.Text, props.Text{
			Size:      10,
			Align:     align.Left,
			Top:       1,
			Hyperlink: &[]string{resume.Contact.Website.URL}[0],
			Color:     &props.Color{Red: 0, Green: 0, Blue: 255},
			Style:     fontstyle.BoldItalic,
		}),
		marotoimg.NewFromFileCol(1, "icons-png/linkedin.png", props.Rect{
			Center:  false,
			Percent: 60,
			Top:     1,
			Left:    10,
		}),
		text.NewCol(3, resume.Contact.LinkedIn.Text, props.Text{
			Size:      10,
			Align:     align.Left,
			Top:       1,
			Hyperlink: &[]string{resume.Contact.LinkedIn.URL}[0],
			Color:     &props.Color{Red: 0, Green: 0, Blue: 255},
			Style:     fontstyle.BoldItalic,
		}),
		marotoimg.NewFromFileCol(1, "icons-png/github.png", props.Rect{
			Center:  false,
			Percent: 60,
			Top:     1,
			Left:    10,
		}),
		text.NewCol(3, resume.Contact.GitHub.Text, props.Text{
			Size:      10,
			Align:     align.Left,
			Top:       1,
			Hyperlink: &[]string{resume.Contact.GitHub.URL}[0],
			Color:     &props.Color{Red: 0, Green: 0, Blue: 255},
			Style:     fontstyle.BoldItalic,
		}),
	)

	// Separator line
	mrt.AddRow(.25, col.New(1)).WithStyle(&props.Cell{BackgroundColor: getPrimaryColor()})

	// Professional title
	mrt.AddRows(text.NewRow(10, resume.Title, props.Text{
		Top:    3,
		Style:  fontstyle.Bold,
		Size:   14,
		Align:  align.Left,
		Family: "dejavu",
	}))

	mrt.AddRows(text.NewRow(15, resume.Summary, props.Text{
		Top:    2,
		Size:   10,
		Align:  align.Left,
		Family: "dejavu-light",
	}))

	// Add spacing after summary
	mrt.AddRow(spaceAfterSummary, col.New(1))

	// Technical Skills
	mrt.AddRow(rowHeightSectionHeader,
		marotoimg.NewFromFileCol(0, "icons-png/box.png", props.Rect{
			Center:  false,
			Percent: iconPercent,
			Left:    iconLeftOffset,
		}),
		text.NewCol(12, "TECHNICAL SKILLS", props.Text{
			Top:   sectionHeaderTopPadding,
			Style: fontstyle.Bold,
			Size:  fontSizeSectionHeader,
			Align: align.Left,
			Color: getPrimaryColor(),
			Left:  sectionHeaderLeftPadding,
		}),
	)

	// Technical Skills underline
	mrt.AddRow(separatorThickness, col.New(1)).WithStyle(&props.Cell{BackgroundColor: getPrimaryColor()})

	// Add each skill category dynamically
	for _, skill := range resume.Skills {
		// Calculate actual line count for skills items (9 cols, 10pt font)
		lines := calculateLineCount(skill.Items, 9, fontSizeSkills)
		skillRowHeight := skillRowBase + float64(lines)*skillRowPerLine

		mrt.AddRow(skillRowHeight,
			text.NewCol(3, skill.Category+":", props.Text{
				Size:  fontSizeSkills,
				Style: fontstyle.Bold,
				Align: align.Left,
			}),
			text.NewCol(9, skill.Items, props.Text{
				Size:  fontSizeSkills,
				Align: align.Left,
			}),
		)
	}

	// Work Experience
	mrt.AddRow(rowHeightSectionHeader,
		marotoimg.NewFromFileCol(0, "icons-png/building-2.png", props.Rect{
			Center:  false,
			Percent: iconPercent,
			Left:    iconLeftOffset,
		}),
		text.NewCol(12, "WORK EXPERIENCE", props.Text{
			Top:   sectionHeaderTopPadding,
			Style: fontstyle.Bold,
			Size:  fontSizeSectionHeader,
			Align: align.Left,
			Color: getPrimaryColor(),
			Left:  sectionHeaderLeftPadding,
		}),
	)

	// underline
	mrt.AddRow(separatorThickness, col.New(1)).WithStyle(&props.Cell{BackgroundColor: getPrimaryColor()})

	for _, exp := range resume.Experience {
		jobTitle := exp.Title + " - " + exp.Company
		if exp.URL != "" {
			mrt.AddRow(rowHeightJobTitle,
				text.NewCol(8, jobTitle, props.Text{
					Size:      fontSizeJobTitle,
					Style:     fontstyle.BoldItalic,
					Align:     align.Left,
					Color:     &props.Color{Red: 0, Green: 0, Blue: 255},
					Hyperlink: &[]string{exp.URL}[0],
				}),
				text.NewCol(4, exp.Dates, props.Text{
					Size:  fontSizeContact,
					Align: align.Right,
				}),
			)
		} else {
			mrt.AddRow(rowHeightJobTitle,
				text.NewCol(8, jobTitle, props.Text{
					Size:  fontSizeJobTitle,
					Style: fontstyle.Bold,
					Align: align.Left,
				}),
				text.NewCol(4, exp.Dates, props.Text{
					Size:  fontSizeContact,
					Align: align.Right,
				}),
			)
		}

		// Achievements
		for _, achievement := range exp.Achievements {
			// Calculate line count for achievements (12 cols, 9pt font, with bullet)
			lines := calculateLineCount("• "+achievement.Text, 12, fontSizeBullet)
			rowHeight := bulletRowBase + float64(lines)*bulletRowPerLine

			mrt.AddRows(text.NewRow(rowHeight, "• "+achievement.Text, props.Text{
				Size:  fontSizeBullet,
				Align: align.Left,
			}))
		}

		// Tech stack
		mrt.AddRows(text.NewRow(rowHeightTechStack, "Tech: "+exp.Tech, props.Text{
			Size:  fontSizeTech,
			Align: align.Left,
			Style: fontstyle.Italic,
		}))
	}

	// Projects (only if present)
	if len(resume.Projects) > 0 {
		mrt.AddRow(rowHeightSectionHeader,
			marotoimg.NewFromFileCol(0, "icons-png/layers.png", props.Rect{
				Center:  false,
				Percent: iconPercent,
				Left:    iconLeftOffset,
			}),
			text.NewCol(12, "PROJECTS", props.Text{
				Top:   sectionHeaderTopPadding,
				Style: fontstyle.Bold,
				Size:  fontSizeSectionHeader,
				Align: align.Left,
				Color: getPrimaryColor(),
				Left:  sectionHeaderLeftPadding,
			}),
		)

		// Projects underline
		mrt.AddRow(separatorThickness, col.New(1)).WithStyle(&props.Cell{BackgroundColor: getPrimaryColor()})

		// Add each project dynamically
		for _, proj := range resume.Projects {
			// Project name with link
			if proj.URL != "" {
				mrt.AddRow(rowHeightJobTitle,
					text.NewCol(12, proj.Name, props.Text{
						Size:      fontSizeJobTitle,
						Style:     fontstyle.BoldItalic,
						Align:     align.Left,
						Color:     &props.Color{Red: 0, Green: 0, Blue: 255},
						Hyperlink: &[]string{proj.URL}[0],
					}),
				)
			} else {
				mrt.AddRow(rowHeightJobTitle,
					text.NewCol(12, proj.Name, props.Text{
						Size:  fontSizeJobTitle,
						Style: fontstyle.Bold,
						Align: align.Left,
					}),
				)
			}

			// Achievements
			for _, achievement := range proj.Achievements {
				// Calculate line count for achievements (12 cols, 9pt font, with bullet)
				lines := calculateLineCount("• "+achievement.Text, 12, fontSizeBullet)
				rowHeight := bulletRowBase + float64(lines)*bulletRowPerLine

				mrt.AddRows(text.NewRow(rowHeight, "• "+achievement.Text, props.Text{
					Size:  fontSizeBullet,
					Align: align.Left,
				}))
			}

			// Tech stack
			mrt.AddRows(text.NewRow(rowHeightTechStack, "Tech: "+proj.Tech, props.Text{
				Size:  fontSizeTech,
				Align: align.Left,
				Style: fontstyle.Italic,
			}))
		}
	}

	// Education
	mrt.AddRow(rowHeightSectionHeader,
		marotoimg.NewFromFileCol(0, "icons-png/graduation-cap.png", props.Rect{
			Center:  false,
			Percent: iconPercent,
			Left:    iconLeftOffset,
		}),
		text.NewCol(12, "EDUCATION", props.Text{
			Top:   sectionHeaderTopPadding,
			Style: fontstyle.Bold,
			Size:  fontSizeSectionHeader,
			Align: align.Left,
			Color: getPrimaryColor(),
			Left:  sectionHeaderLeftPadding,
		}),
	)

	// Education underline
	mrt.AddRow(separatorThickness, col.New(1)).WithStyle(&props.Cell{BackgroundColor: getPrimaryColor()})

	// Add each education entry dynamically
	for _, edu := range resume.Education {
		if edu.URL != "" {
			mrt.AddRow(8,
				text.NewCol(8, edu.School, props.Text{
					Size:      11,
					Style:     fontstyle.BoldItalic,
					Align:     align.Left,
					Color:     &props.Color{Red: 0, Green: 0, Blue: 255},
					Hyperlink: &[]string{edu.URL}[0],
				}),
				text.NewCol(4, edu.Date, props.Text{
					Size:  10,
					Align: align.Right,
				}),
			)
		} else {
			mrt.AddRow(8,
				text.NewCol(8, edu.School, props.Text{
					Size:  11,
					Style: fontstyle.Bold,
					Align: align.Left,
				}),
				text.NewCol(4, edu.Date, props.Text{
					Size:  10,
					Align: align.Right,
				}),
			)
		}

		// Degree
		mrt.AddRows(text.NewRow(8, edu.Degree, props.Text{
			Size:  10,
			Align: align.Left,
		}))
	}

	return mrt
}

func getPrimaryColor() *props.Color {
	return &props.Color{
		Red:   70,
		Green: 130,
		Blue:  180,
	}
}
