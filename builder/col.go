package builder

// ColBuilder represents a single column cell in a 12-column grid row.
// It embeds Container, giving it the full set of content methods:
// Text, Row, AutoRow, Image, Line, Spacer, PageBreak, KeepTogether,
// EnsureSpace, and PageNumber.
//
// ColBuilder instances are created by RowBuilder.Col and must not be
// constructed directly.
type ColBuilder struct {
	// Container provides all content-adding methods.
	Container
}
