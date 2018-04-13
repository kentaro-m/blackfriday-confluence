package confluence_test

import (
	"testing"

	bfconfluence "github.com/kentaro-m/blackfriday-confluence"
	bf "gopkg.in/russross/blackfriday.v2"
)

type testData struct {
	input      string
	expected   string
	extensions bf.Extensions
}

func doTest(t *testing.T, tdt []testData) {
	for _, v := range tdt {
		renderer := &bfconfluence.Renderer{}
		md := bf.New(bf.WithRenderer(renderer), bf.WithExtensions(v.extensions))
		ast := md.Parse([]byte(v.input))
		output := string(renderer.Render(ast))

		if output != v.expected {
			t.Errorf("got:%v\nwant:%v", output, v.expected)
		}
	}
}

func TestHeading(t *testing.T) {
	tdt := []testData{
		{input: "# Section\n", expected: "h1. Section\n", extensions: bf.CommonExtensions},
		{input: "## SubSection\n", expected: "h2. SubSection\n", extensions: bf.CommonExtensions},
		{input: "### SubSubSection\n", expected: "h3. SubSubSection\n", extensions: bf.CommonExtensions},
	}

	doTest(t, tdt)
}

func TestBlockQuote(t *testing.T) {
	tdt := []testData{
		{
			input:      "> block quote",
			expected:   "{quote}\nblock quote\n\n{quote}\n\n",
			extensions: bf.CommonExtensions,
		},
		{
			input:      "> block quote\n> block quote",
			expected:   "{quote}\nblock quote\nblock quote\n\n{quote}\n\n",
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestCodeBlock(t *testing.T) {
	tdt := []testData{
		{
			input:      "```c\n\nint main(void) {\n printf(\"Hello, world.\"); \n}\n```",
			expected:   "{code}\n\nint main(void) {\n printf(\"Hello, world.\"); \n}\n{code}\n\n",
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestImage(t *testing.T) {
	tdt := []testData{
		{
			input:      "![](./sample.png)",
			expected:   "!./sample.png!\n\n",
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestList(t *testing.T) {
	tdt := []testData{
		{
			input:      "* list1\n* list2\n* list 3\n",
			expected:   "* list1\n* list2\n* list 3\n",
			extensions: bf.CommonExtensions,
		},
		{
			input:      "* list1\n* list2\n  * list 3\n  * list 4\n* list 5\n",
			expected:   "* list1\n* list2\n** list 3\n** list 4\n* list 5\n",
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestOrderedList(t *testing.T) {
	tdt := []testData{
		{
			input:      "1. list1\n1. list2\n1. list3\n",
			expected:   "# list1\n# list2\n# list3\n",
			extensions: bf.CommonExtensions,
		},
		{
			input:      "1. list1\n  1. list2\n1. list3\n",
			expected:   "# list1\n## list2\n# list3\n",
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestLink(t *testing.T) {
	tdt := []testData{
		{
			input:      "[Example Domain](http://www.example.com/)",
			expected:   "[Example Domain|http://www.example.com/]\n\n",
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestHorizontalRule(t *testing.T) {
	tdt := []testData{
		{
			input:      "- - - -",
			expected:   "----\n",
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestStrong(t *testing.T) {
	tdt := []testData{
		{
			input:      "**strong text**",
			expected:   "*strong text*\n\n",
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestEmph(t *testing.T) {
	tdt := []testData{
		{
			input:      "_emph text_",
			expected:   "_emph text_\n\n",
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestDel(t *testing.T) {
	tdt := []testData{
		{
			input:      "~~del text~~",
			expected:   "-del text-\n\n",
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestTable(t *testing.T) {
	tdt := []testData{
		{
			input: `
First Header  | Second Header
------------- | -------------
Content Cell  | Content Cell
Content Cell  | Content Cell`,
			expected: `||First Header||Second Header||
|Content Cell|Content Cell|
|Content Cell|Content Cell|

`,
			extensions: bf.CommonExtensions,
		},
		{
			input: `
|a  |b  |c  |
|---|---|---|
|1  |2  |3  |
|4  |5  |6  |`,
			expected: `||a||b||c||
|1|2|3|
|4|5|6|

`,
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestEsc(t *testing.T) {
	tdt := []testData{
		{
			input:      "*-_+",
			expected:   "\\*\\-\\_\\+",
			extensions: bf.CommonExtensions,
		},
	}

	doTest(t, tdt)
}

func TestRun(t *testing.T) {
	input := `
# Section
hello, world.
`
	expected := `h1. Section
hello, world.

`

	output := string(bfconfluence.Run([]byte(input)))

	if output != expected {
		t.Errorf("got:%v\nwant:%v", output, expected)
	}
}
