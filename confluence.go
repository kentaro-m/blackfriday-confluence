package confluence

import (
	"bytes"
	"io"

	bf "github.com/russross/blackfriday/v2"
)

// Renderer is the rendering interface for confluence wiki output.
type Renderer struct {
	w bytes.Buffer

	// Flags allow customizing this renderer's behavior.
	Flags Flag

	lastOutputLen int
}

// Flag control optional behavior of this renderer.
type Flag int

const (
	// FlagsNone does not allow customizing this renderer's behavior.
	FlagsNone Flag = 0

	// InformationMacros allow using info, tip, note, and warning macros.
	InformationMacros Flag = 1 << iota

	// IgnoreMacroEscaping will not escape any text that contains starts with `{`
	// in a block of text.
	IgnoreMacroEscaping
)

var (
	quoteTag         = []byte("{quote}")
	codeTag          = []byte("code")
	infoTag          = []byte("info")
	noteTag          = []byte("note")
	tipTag           = []byte("tip")
	warningTag       = []byte("warning")
	imageTag         = []byte("!")
	strongTag        = []byte("*")
	strikethroughTag = []byte("-")
	emTag            = []byte("_")
	linkTag          = []byte("[")
	linkCloseTag     = []byte("]")
	liTag            = []byte("*")
	olTag            = []byte("#")
	hrTag            = []byte("----")
	tableTag         = []byte("|")
	h1Tag            = []byte("h1.")
	h2Tag            = []byte("h2.")
	h3Tag            = []byte("h3.")
	h4Tag            = []byte("h4.")
	h5Tag            = []byte("h5.")
	h6Tag            = []byte("h6.")
)

var (
	nlBytes    = []byte{'\n'}
	spaceBytes = []byte{' '}
)

var itemLevel = 0

var confluenceEscaper = [256][]byte{
	'*': []byte(`\*`),
	'_': []byte(`\_`),
	'-': []byte(`\-`),
	'+': []byte(`\+`),
	'^': []byte(`\^`),
	'~': []byte(`\~`),
	'{': []byte(`\{`),
	'!': []byte(`\!`),
	'[': []byte(`\[`),
	']': []byte(`\]`),
	'(': []byte(`\(`),
	')': []byte(`\)`),
}

func (r *Renderer) esc(w io.Writer, text []byte) {
	var start, end int
	for end < len(text) {
		if escSeq := confluenceEscaper[text[end]]; escSeq != nil {

			// do not escape blocks that start with `{` if request.
			// this will allow people to declare macros in `wiki` format.
			if r.Flags&IgnoreMacroEscaping != 0 && text[end] == '{' {
				escSeq = escSeq[1:]
			}

			w.Write(text[start:end])
			w.Write(escSeq)
			start = end + 1
		}
		end++
	}

	if start < len(text) && end <= len(text) {
		w.Write(text[start:end])
	}
}

func (r *Renderer) cr(w io.Writer) {
	if r.lastOutputLen > 0 {
		r.out(w, nlBytes)
	}
}

func (r *Renderer) out(w io.Writer, text []byte) {
	w.Write(text)
	r.lastOutputLen = len(text)
}

func headingTagFromLevel(level int) []byte {
	switch level {
	case 1:
		return h1Tag
	case 2:
		return h2Tag
	case 3:
		return h3Tag
	case 4:
		return h4Tag
	case 5:
		return h5Tag
	default:
		return h6Tag
	}
}

// RenderNode is a confluence renderer of a single node of a syntax tree.
func (r *Renderer) RenderNode(w io.Writer, node *bf.Node, entering bool) bf.WalkStatus {
	switch node.Type {
	case bf.Text:
		r.esc(w, node.Literal)
	case bf.Softbreak:
		break
	case bf.Hardbreak:
		w.Write(nlBytes)
	case bf.BlockQuote:
		if entering {
			r.out(w, quoteTag)
			r.cr(w)
		} else {
			r.out(w, quoteTag)
			r.cr(w)
			r.cr(w)
		}
	case bf.CodeBlock:
		r.out(w, []byte("{"))
		if len(node.Info) > 0 {
			if r.Flags&InformationMacros != 0 {
				language := string(node.Info)
				switch language {
				case "info":
					r.out(w, infoTag)
				case "tip":
					r.out(w, tipTag)
				case "note":
					r.out(w, noteTag)
				case "warning":
					r.out(w, warningTag)
				default:
					r.out(w, []byte(codeTag))
					r.out(w, []byte(":"))
					r.out(w, node.Info)
				}
				r.out(w, []byte("}"))
				r.cr(w)
				w.Write(node.Literal)
				r.out(w, []byte("{"))
				switch language {
				case "info":
					r.out(w, infoTag)
				case "tip":
					r.out(w, tipTag)
				case "note":
					r.out(w, noteTag)
				case "warning":
					r.out(w, warningTag)
				default:
					r.out(w, []byte(codeTag))
				}
			} else {
				r.out(w, []byte(codeTag))
				r.out(w, []byte(":"))
				r.out(w, node.Info)
				r.out(w, []byte("}"))
				r.cr(w)
				w.Write(node.Literal)
				r.out(w, []byte("{"))
				r.out(w, []byte(codeTag))
			}
		} else {
			r.out(w, codeTag)
			r.out(w, []byte("}"))
			r.cr(w)
			w.Write(node.Literal)
			r.out(w, []byte("{"))
			r.out(w, codeTag)
		}
		r.out(w, []byte("}"))
		r.cr(w)
		r.cr(w)
	case bf.Code:
		r.out(w, []byte("{{"))
		r.esc(w, node.Literal)
		r.out(w, []byte("}}"))
	case bf.Emph:
		r.out(w, emTag)
	case bf.Heading:
		headingTag := headingTagFromLevel(node.Level)
		if entering {
			r.out(w, headingTag)
			w.Write(spaceBytes)
		} else {
			r.cr(w)
		}
	case bf.Image:
		if entering {
			dest := node.LinkData.Destination
			r.out(w, imageTag)
			r.out(w, dest)
		} else {
			r.out(w, imageTag)
		}
	case bf.Item:
		if entering {
			itemTag := liTag
			if node.ListFlags&bf.ListTypeOrdered != 0 {
				itemTag = olTag
			}

			for i := 0; i < itemLevel; i++ {
				r.out(w, itemTag)
			}

			w.Write(spaceBytes)
		}
	case bf.Link:
		if entering {
			r.out(w, linkTag)
		} else {
			if dest := node.LinkData.Destination; dest != nil {
				r.out(w, []byte("|"))
				r.out(w, dest)
			}
			r.out(w, linkCloseTag)
		}
	case bf.HorizontalRule:
		r.cr(w)
		r.out(w, hrTag)
		r.cr(w)
	case bf.List:
		if entering {
			itemLevel++
		} else {
			itemLevel--
			if itemLevel == 0 {
				r.cr(w)
			}
		}
	case bf.Document:
		break
	case bf.HTMLBlock:
		break
	case bf.HTMLSpan:
		break
	case bf.Paragraph:
		if !entering {
			if node.Next != nil && node.Next.Type == bf.Paragraph {
				w.Write(nlBytes)
				w.Write(nlBytes)
			} else {
				if node.Parent.Type != bf.Item {
					r.cr(w)
				}
				r.cr(w)
			}
		}
	case bf.Strong:
		r.out(w, strongTag)
	case bf.Del:
		r.out(w, strikethroughTag)
	case bf.Table:
		if !entering {
			r.cr(w)
		}
	case bf.TableCell:
		if node.IsHeader {
			r.out(w, tableTag)
		} else {
			if entering {
				r.out(w, tableTag)
			}
		}
	case bf.TableHead:
		break
	case bf.TableBody:
		break
	case bf.TableRow:
		if node.Parent.Type == bf.TableHead {
			r.out(w, tableTag)

			if !entering {
				r.cr(w)
			}
		} else if node.Parent.Type == bf.TableBody {
			if !entering {
				r.out(w, tableTag)
				r.cr(w)
			}
		}
	default:
		panic("Unknown node type " + node.Type.String())
	}
	return bf.GoToNext
}

// Render prints out the whole document from the ast.
func (r *Renderer) Render(ast *bf.Node) []byte {
	ast.Walk(func(node *bf.Node, entering bool) bf.WalkStatus {
		return r.RenderNode(&r.w, node, entering)
	})

	return r.w.Bytes()
}

// RenderHeader writes document header (unused).
func (r *Renderer) RenderHeader(w io.Writer, ast *bf.Node) {
}

// RenderFooter writes document footer (unused).
func (r *Renderer) RenderFooter(w io.Writer, ast *bf.Node) {
}

// Run prints out the confluence document.
func Run(input []byte, opts ...bf.Option) []byte {
	r := &Renderer{Flags: InformationMacros}
	optList := []bf.Option{bf.WithRenderer(r), bf.WithExtensions(bf.CommonExtensions)}
	optList = append(optList, opts...)
	parser := bf.New(optList...)
	ast := parser.Parse([]byte(input))
	return r.Render(ast)
}
