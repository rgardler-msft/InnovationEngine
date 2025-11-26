package parsers

import (
	"fmt"
	"regexp"
	"testing"
)

func TestParsingMarkdownHeaders(t *testing.T) {
	t.Run("Markdown with a valid title", func(t *testing.T) {
		markdown := []byte(`# Hello World`)
		document := ParseMarkdownIntoAst(markdown)
		title, err := ExtractScenarioTitleFromAst(document, markdown)
		if err != nil {
			t.Errorf("Error parsing title: %s", err)
		}

		if title != "Hello World" {
			t.Errorf("Title is wrong: %s", title)
		}
	})

	t.Run("Markdown with multiple titles", func(t *testing.T) {
		markdown := []byte("# Hello World \n # Hello again")
		document := ParseMarkdownIntoAst(markdown)
		title, err := ExtractScenarioTitleFromAst(document, markdown)
		if err != nil {
			t.Errorf("Error parsing title: %s", err)
		}

		if title != "Hello World" {
			t.Errorf("Title is wrong: %s", title)
		}
	})

	t.Run("Markdown without a title", func(t *testing.T) {
		markdown := []byte(``)

		document := ParseMarkdownIntoAst(markdown)
		title, err := ExtractScenarioTitleFromAst(document, markdown)

		if err == nil {
			t.Errorf("Error should have been thrown")
		}

		if title != "" {
			t.Errorf("Title should be empty")
		}
	})
}

func TestParsingYamlMetadata(t *testing.T) {
	t.Run("Markdown with valid yaml metadata", func(t *testing.T) {
		markdown := []byte(`---
    key: value
    array: [1, 2, 3]
    ---
    `)

		document := ParseMarkdownIntoAst(markdown)
		metadata := ExtractYamlMetadataFromAst(document)

		if metadata["key"] != "value" {
			t.Errorf("Metadata is wrong: %v", metadata)
		}

		array := metadata["array"].([]interface{})
		if array[0] != 1 || array[1] != 2 || array[2] != 3 {
			t.Errorf("Metadata is wrong: %v", metadata)
		}
	})

	t.Run("Markdown without yaml metadata", func(t *testing.T) {
		markdown := []byte(`# Hello World.`)
		document := ParseMarkdownIntoAst(markdown)
		metadata := ExtractYamlMetadataFromAst(document)

		if len(metadata) != 0 {
			t.Errorf("Metadata should be empty")
		}
	})

	t.Run("yaml with nested properties", func(t *testing.T) {
		markdown := []byte(`---
    nested:
      key: value
    key.value: otherValue
    ---
    `)

		document := ParseMarkdownIntoAst(markdown)
		metadata := ExtractYamlMetadataFromAst(document)

		nested := metadata["nested"].(map[interface{}]interface{})
		if nested["key"] != "value" {
			t.Errorf("Metadata is wrong: %v", metadata)
		}

		if metadata["key.value"] != "otherValue" {
			t.Errorf("Metadata is wrong: %v", metadata)
		}
	})
}

func TestParsingMarkdownCodeBlocks(t *testing.T) {
	t.Run("Markdown with a valid bash code block", func(t *testing.T) {
		markdown := []byte(fmt.Sprintf("# Hello World\n ```bash\n%s\n```", "echo Hello"))

		document := ParseMarkdownIntoAst(markdown)
		codeBlocks := ExtractCodeBlocksFromAst(document, markdown, []string{"bash"}, "test.md")

		if len(codeBlocks) != 1 {
			t.Errorf("Code block count is wrong: %d", len(codeBlocks))
		}

		if codeBlocks[0].Language != "bash" {
			t.Errorf("Code block language is wrong: %s", codeBlocks[0].Language)
		}

		if codeBlocks[0].Content != "echo Hello\n" {
			t.Errorf(
				"Code block code is wrong. Expected: %s, Got %s",
				"echo Hello\\n",
				codeBlocks[0].Content,
			)
		}
	})
}

func TestParsingMarkdownExpectedSimilarty(t *testing.T) {
	t.Run("Markdown with a expected_similarty tag using float", func(t *testing.T) {
		markdown := []byte(
			fmt.Sprintf(
				"```bash\n%s\n```\n<!--expected_similarity=0.8-->\n```\nHello\n```\n",
				"echo Hello",
			),
		)

		document := ParseMarkdownIntoAst(markdown)
		codeBlocks := ExtractCodeBlocksFromAst(document, markdown, []string{"bash"}, "test.md")

		if len(codeBlocks) != 1 {
			t.Errorf("Code block count is wrong: %d", len(codeBlocks))
		}

		block := codeBlocks[0].ExpectedOutput
		expectedFloat := .8
		if block.ExpectedSimilarity != expectedFloat {
			t.Errorf(
				"ExpectedSimilarity is wrong, got %f, expected %f",
				block.ExpectedSimilarity,
				expectedFloat,
			)
		}
	})
}

func TestParsingMarkdownExpectedRegex(t *testing.T) {
	t.Run("Markdown with a expected_similarty tag using regex", func(t *testing.T) {
		markdown := []byte(
			fmt.Sprintf(
				"```bash\n%s\n```\n<!--expected_similarity=\"Foo \\w+\"-->\n```\nFoo Bar\n```\n",
				"echo 'Foo Bar'",
			),
		)

		document := ParseMarkdownIntoAst(markdown)
		codeBlocks := ExtractCodeBlocksFromAst(document, markdown, []string{"bash"}, "test.md")

		if len(codeBlocks) != 1 {
			t.Errorf("Code block count is wrong: %d", len(codeBlocks))
		}

		block := codeBlocks[0].ExpectedOutput
		if block.ExpectedRegex == nil {
			t.Errorf("ExpectedRegex is nil")
		}

		stringRegex := block.ExpectedRegex.String()
		expectedRegex := `Foo \w+`
		if stringRegex != expectedRegex {
			t.Errorf("ExpectedRegex is wrong, got %q, expected %q", stringRegex, expectedRegex)
		}
	})
}

func TestExtractSectionTextFromMarkdown(t *testing.T) {
	markdown := []byte(`# Title

Intro paragraph.

## Prerequisites

This is the prerequisites section.

- Item one
- Item two

## Next Section

More text.
`)

	section := ExtractSectionTextFromMarkdown(markdown, "Prerequisites")

	expectedPattern := `(?s)^This is the prerequisites section\.\s+- Item one\s+- Item two$`
	if match, _ := regexp.MatchString(expectedPattern, section); !match {
		t.Fatalf("Section text did not match expected content. Got: %q", section)
	}
}

func TestCodeBlocksInsidePrerequisiteSectionAreMarked(t *testing.T) {
	markdown := []byte("# Title\n\nIntro paragraph.\n\n## Prerequisite\n\n### First Action\n\n```bash\necho \"first\"\n```\n\n### Second Action\n\n```bash\necho \"second\"\n```\n\n## Steps\n\n```bash\necho \"step\"\n```\n")

	document := ParseMarkdownIntoAst(markdown)
	blocks := ExtractCodeBlocksFromAst(document, markdown, []string{"bash"}, "test.md")
	if len(blocks) != 3 {
		t.Fatalf("expected 3 code blocks, got %d", len(blocks))
	}
	if !blocks[0].InPrerequisiteSection || !blocks[1].InPrerequisiteSection {
		t.Fatalf("expected prerequisite blocks to be marked as such")
	}
	if blocks[2].InPrerequisiteSection {
		t.Fatalf("non-prerequisite blocks should not be marked as prerequisites")
	}
}

func TestExtractPrerequisiteUrlsHandlesSingularHeading(t *testing.T) {
	markdown := []byte("# Title\n\n## Prerequisite\n\n- [First](first.md)\n- [Second](second.md)\n\n## Steps\n\n```bash\necho \"done\"\n```\n")

	document := ParseMarkdownIntoAst(markdown)
	urls, err := ExtractPrerequisiteUrlsFromAst(document, markdown)
	if err != nil {
		t.Fatalf("unexpected error extracting urls: %v", err)
	}
	if len(urls) != 2 {
		t.Fatalf("expected 2 urls, got %d", len(urls))
	}
	if urls[0] != "first.md" || urls[1] != "second.md" {
		t.Fatalf("urls not preserved in order: %#v", urls)
	}
}
