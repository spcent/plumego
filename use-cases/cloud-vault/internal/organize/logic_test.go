package organize

import (
	"context"
	"strings"
	"testing"

	"cloud-vault/internal/document"
)

// ---------------------------------------------------------------------------
// normalizeTitle
// ---------------------------------------------------------------------------

func TestNormalizeTitle_Basic(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Hello World", "hello world"},
		{"  Leading Trailing  ", "leading trailing"},
		// The regex collapses consecutive non-alpha chars into a single space.
		{"Hello, World! 123", "hello world 123"},
		{"", ""},
		{"ALL CAPS", "all caps"},
		{"foo-bar_baz", "foo bar baz"},
		{"日本語テスト", "日本語テスト"},
		// Multiple spaces are collapsed.
		{"Multiple   Spaces", "multiple spaces"},
	}
	for _, tc := range tests {
		got := normalizeTitle(tc.in)
		if got != tc.want {
			t.Errorf("normalizeTitle(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeTitle_Lowercase(t *testing.T) {
	got := normalizeTitle("UPPER lower MiXeD")
	if got != strings.ToLower("UPPER lower MiXeD") {
		// normalizeTitle lowercases and strips non-alpha
		if strings.ToUpper(got) == got {
			t.Errorf("expected lowercase output, got %q", got)
		}
	}
}

// ---------------------------------------------------------------------------
// tokenize
// ---------------------------------------------------------------------------

func TestTokenize_Basic(t *testing.T) {
	tokens := tokenize("Hello World foo")
	if len(tokens) == 0 {
		t.Fatal("expected tokens, got none")
	}
	for _, tok := range tokens {
		if len(tok) < 2 {
			t.Errorf("token %q is shorter than 2 chars", tok)
		}
	}
}

func TestTokenize_Lowercase(t *testing.T) {
	tokens := tokenize("Hello World")
	for _, tok := range tokens {
		if tok != strings.ToLower(tok) {
			t.Errorf("token %q is not lowercase", tok)
		}
	}
}

func TestTokenize_FiltersShortTokens(t *testing.T) {
	// Single-char tokens should be excluded.
	tokens := tokenize("a b c hello world")
	for _, tok := range tokens {
		if len(tok) < 2 {
			t.Errorf("short token %q should have been filtered", tok)
		}
	}
}

func TestTokenize_EmptyString(t *testing.T) {
	tokens := tokenize("")
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for empty string, got %d", len(tokens))
	}
}

func TestTokenize_OnlyPunctuation(t *testing.T) {
	tokens := tokenize("!!! ??? ,,, ---")
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens for only-punctuation string, got %d", len(tokens))
	}
}

// ---------------------------------------------------------------------------
// buildShingles
// ---------------------------------------------------------------------------

func TestBuildShingles_Empty(t *testing.T) {
	shingles := buildShingles([]string{})
	if len(shingles) != 0 {
		t.Errorf("expected empty shingles, got %d", len(shingles))
	}
}

func TestBuildShingles_SingleToken(t *testing.T) {
	shingles := buildShingles([]string{"hello"})
	if _, ok := shingles["hello"]; !ok {
		t.Error("expected unigram 'hello' in shingles")
	}
	// No bigram possible with one token.
	for k := range shingles {
		if strings.Contains(k, " ") {
			t.Errorf("unexpected bigram %q with single token", k)
		}
	}
}

func TestBuildShingles_TwoTokens(t *testing.T) {
	shingles := buildShingles([]string{"foo", "bar"})
	if _, ok := shingles["foo bar"]; !ok {
		t.Error("expected bigram 'foo bar'")
	}
	if _, ok := shingles["foo"]; !ok {
		t.Error("expected unigram 'foo'")
	}
	if _, ok := shingles["bar"]; !ok {
		t.Error("expected unigram 'bar'")
	}
}

func TestBuildShingles_MultipleBigrams(t *testing.T) {
	shingles := buildShingles([]string{"a", "b", "c"})
	expected := []string{"a b", "b c", "a", "b", "c"}
	for _, e := range expected {
		if _, ok := shingles[e]; !ok {
			t.Errorf("expected shingle %q not found", e)
		}
	}
}

// ---------------------------------------------------------------------------
// jaccardSimilarity
// ---------------------------------------------------------------------------

func TestJaccardSimilarity_IdenticalSets(t *testing.T) {
	a := map[string]struct{}{"foo": {}, "bar": {}, "baz": {}}
	b := map[string]struct{}{"foo": {}, "bar": {}, "baz": {}}
	got := jaccardSimilarity(a, b)
	if got != 1.0 {
		t.Errorf("jaccardSimilarity identical sets = %f, want 1.0", got)
	}
}

func TestJaccardSimilarity_DisjointSets(t *testing.T) {
	a := map[string]struct{}{"foo": {}}
	b := map[string]struct{}{"bar": {}}
	got := jaccardSimilarity(a, b)
	if got != 0.0 {
		t.Errorf("jaccardSimilarity disjoint sets = %f, want 0.0", got)
	}
}

func TestJaccardSimilarity_BothEmpty(t *testing.T) {
	got := jaccardSimilarity(map[string]struct{}{}, map[string]struct{}{})
	if got != 1.0 {
		t.Errorf("jaccardSimilarity both empty = %f, want 1.0", got)
	}
}

func TestJaccardSimilarity_OneEmpty(t *testing.T) {
	a := map[string]struct{}{"foo": {}}
	b := map[string]struct{}{}
	got := jaccardSimilarity(a, b)
	if got != 0.0 {
		t.Errorf("jaccardSimilarity one empty = %f, want 0.0", got)
	}
}

func TestJaccardSimilarity_PartialOverlap(t *testing.T) {
	a := map[string]struct{}{"foo": {}, "bar": {}}
	b := map[string]struct{}{"foo": {}, "baz": {}}
	got := jaccardSimilarity(a, b)
	// intersection=1, union=3 → 1/3 ≈ 0.333
	expected := 1.0 / 3.0
	if diff := got - expected; diff > 0.001 || diff < -0.001 {
		t.Errorf("jaccardSimilarity partial = %f, want ~%f", got, expected)
	}
}

// ---------------------------------------------------------------------------
// simhashFromTokens
// ---------------------------------------------------------------------------

func TestSimhashFromTokens_Empty(t *testing.T) {
	got := simhashFromTokens([]string{})
	if len(got) != 16 {
		t.Errorf("simhash length = %d, want 16 hex chars", len(got))
	}
}

func TestSimhashFromTokens_Deterministic(t *testing.T) {
	tokens := []string{"hello", "world", "test"}
	a := simhashFromTokens(tokens)
	b := simhashFromTokens(tokens)
	if a != b {
		t.Errorf("simhash not deterministic: %q vs %q", a, b)
	}
}

func TestSimhashFromTokens_DifferentInputs(t *testing.T) {
	a := simhashFromTokens([]string{"hello", "world"})
	b := simhashFromTokens([]string{"foo", "bar"})
	if a == b {
		t.Errorf("expected different simhashes for different inputs, both = %q", a)
	}
}

func TestSimhashFromTokens_HexFormat(t *testing.T) {
	got := simhashFromTokens([]string{"test"})
	for _, c := range got {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("simhash %q contains non-hex char %q", got, c)
		}
	}
}

// ---------------------------------------------------------------------------
// hamming64
// ---------------------------------------------------------------------------

func TestHamming64_EqualStrings(t *testing.T) {
	h := "0123456789abcdef"
	got := hamming64(h, h)
	if got != 0 {
		t.Errorf("hamming64 equal strings = %d, want 0", got)
	}
}

func TestHamming64_AllBitsDifferent(t *testing.T) {
	// 0000...0 vs ffff...f → all 64 bits differ
	got := hamming64("0000000000000000", "ffffffffffffffff")
	if got != 64 {
		t.Errorf("hamming64 all bits different = %d, want 64", got)
	}
}

func TestHamming64_OneBitDifferent(t *testing.T) {
	// 0x0000000000000001 vs 0x0000000000000000 → 1 bit different
	got := hamming64("0000000000000001", "0000000000000000")
	if got != 1 {
		t.Errorf("hamming64 one bit different = %d, want 1", got)
	}
}

func TestHamming64_Symmetric(t *testing.T) {
	a := "0123456789abcdef"
	b := "fedcba9876543210"
	if hamming64(a, b) != hamming64(b, a) {
		t.Error("hamming64 should be symmetric")
	}
}

// ---------------------------------------------------------------------------
// simhashSimilarity
// ---------------------------------------------------------------------------

func TestSimhashSimilarity_Equal(t *testing.T) {
	h := "0000000000000000"
	got := simhashSimilarity(h, h)
	if got != 1.0 {
		t.Errorf("simhashSimilarity equal = %f, want 1.0", got)
	}
}

func TestSimhashSimilarity_AllBitsDifferent(t *testing.T) {
	got := simhashSimilarity("0000000000000000", "ffffffffffffffff")
	if got != 0.0 {
		t.Errorf("simhashSimilarity all bits different = %f, want 0.0", got)
	}
}

func TestSimhashSimilarity_RangeValid(t *testing.T) {
	tokens1 := []string{"hello", "world"}
	tokens2 := []string{"hello", "earth"}
	h1 := simhashFromTokens(tokens1)
	h2 := simhashFromTokens(tokens2)
	sim := simhashSimilarity(h1, h2)
	if sim < 0.0 || sim > 1.0 {
		t.Errorf("simhashSimilarity out of range [0,1]: %f", sim)
	}
}

// ---------------------------------------------------------------------------
// buildBuckets
// ---------------------------------------------------------------------------

func TestBuildBuckets_Empty(t *testing.T) {
	buckets := buildBuckets([]docFingerprintRow{}, map[string][]string{})
	if len(buckets) != 0 {
		t.Errorf("expected 0 buckets for empty docs, got %d", len(buckets))
	}
}

func TestBuildBuckets_TagBucket(t *testing.T) {
	docs := []docFingerprintRow{
		{ID: "doc1", Title: "Alpha"},
		{ID: "doc2", Title: "Beta"},
	}
	docTags := map[string][]string{
		"doc1": {"tag-x"},
		"doc2": {"tag-x"},
	}
	buckets := buildBuckets(docs, docTags)
	found := false
	for _, b := range buckets {
		if b.key == "tag:tag-x" {
			found = true
			if len(b.docs) != 2 {
				t.Errorf("tag bucket docs = %d, want 2", len(b.docs))
			}
		}
	}
	if !found {
		t.Error("expected tag:tag-x bucket")
	}
}

func TestBuildBuckets_JobBucket(t *testing.T) {
	docs := []docFingerprintRow{
		{ID: "doc1", ImportJobID: "job-1"},
		{ID: "doc2", ImportJobID: "job-1"},
	}
	buckets := buildBuckets(docs, map[string][]string{})
	found := false
	for _, b := range buckets {
		if b.key == "job:job-1" {
			found = true
		}
	}
	if !found {
		t.Error("expected job:job-1 bucket")
	}
}

func TestBuildBuckets_DirBucket(t *testing.T) {
	docs := []docFingerprintRow{
		{ID: "doc1", OriginalPath: "notes/work/file1.md"},
		{ID: "doc2", OriginalPath: "notes/work/file2.md"},
	}
	buckets := buildBuckets(docs, map[string][]string{})
	found := false
	for _, b := range buckets {
		if strings.HasPrefix(b.key, "dir:") {
			found = true
		}
	}
	if !found {
		t.Error("expected dir: bucket for docs in same directory")
	}
}

func TestBuildBuckets_PrefixBucket(t *testing.T) {
	docs := []docFingerprintRow{
		{ID: "doc1", Title: "Go Programming Guide Advanced"},
		{ID: "doc2", Title: "Go Programming Guide Basics"},
	}
	buckets := buildBuckets(docs, map[string][]string{})
	found := false
	for _, b := range buckets {
		if strings.HasPrefix(b.key, "prefix:") {
			found = true
		}
	}
	if !found {
		t.Error("expected prefix: bucket for titles sharing 3-word prefix")
	}
}

func TestBuildBuckets_SingleDocPerTag_NoBucket(t *testing.T) {
	// Only one doc per tag → no bucket should form (need ≥2).
	docs := []docFingerprintRow{
		{ID: "doc1", Title: "Only One"},
	}
	docTags := map[string][]string{"doc1": {"tag-solo"}}
	buckets := buildBuckets(docs, docTags)
	for _, b := range buckets {
		if b.key == "tag:tag-solo" {
			t.Error("single-doc tag should not create a bucket")
		}
	}
}

// ---------------------------------------------------------------------------
// computeQualityScore
// ---------------------------------------------------------------------------

func TestComputeQualityScore_Baseline(t *testing.T) {
	row := docScoringRow{
		ID:    "doc1",
		Title: "A Medium Length Title Here",
		// word count 30–99, no headings, not favorite
		WordCount: 50,
	}
	score := computeQualityScore(row)
	if score < 0 || score > 100 {
		t.Errorf("score out of range: %f", score)
	}
}

func TestComputeQualityScore_ShortTitle(t *testing.T) {
	row := docScoringRow{Title: "Hi", WordCount: 200}
	score := computeQualityScore(row)
	baseline := docScoringRow{Title: "Long Enough Title", WordCount: 200}
	scoreBaseline := computeQualityScore(baseline)
	if score >= scoreBaseline {
		t.Errorf("short title should score lower: %f >= %f", score, scoreBaseline)
	}
}

func TestComputeQualityScore_LongTitle(t *testing.T) {
	rowShort := docScoringRow{Title: "Hi", WordCount: 200}
	rowLong := docScoringRow{Title: "This Is A Sufficiently Long Title", WordCount: 200}
	if computeQualityScore(rowLong) <= computeQualityScore(rowShort) {
		t.Error("long title should score higher than short title")
	}
}

func TestComputeQualityScore_HighWordCount(t *testing.T) {
	rowLow := docScoringRow{Title: "Title", WordCount: 10}
	rowHigh := docScoringRow{Title: "Title", WordCount: 500}
	if computeQualityScore(rowHigh) <= computeQualityScore(rowLow) {
		t.Error("high word count should score higher")
	}
}

func TestComputeQualityScore_WithHeadings(t *testing.T) {
	rowNoHead := docScoringRow{Title: "Title", WordCount: 200}
	rowWithHead := docScoringRow{
		Title:        "Title",
		WordCount:    200,
		HeadingsJSON: `[{"text":"Intro"},{"text":"Details"},{"text":"Summary"}]`,
	}
	if computeQualityScore(rowWithHead) <= computeQualityScore(rowNoHead) {
		t.Error("document with headings should score higher")
	}
}

func TestComputeQualityScore_CodeBlocks(t *testing.T) {
	rowNoCode := docScoringRow{Title: "Title", WordCount: 200}
	rowWithCode := docScoringRow{Title: "Title", WordCount: 200, CodeBlockCount: 2}
	if computeQualityScore(rowWithCode) <= computeQualityScore(rowNoCode) {
		t.Error("document with code blocks should score higher")
	}
}

func TestComputeQualityScore_FavoriteBoost(t *testing.T) {
	rowNormal := docScoringRow{Title: "Title", WordCount: 200}
	rowFav := docScoringRow{Title: "Title", WordCount: 200, IsFavorite: true}
	if computeQualityScore(rowFav) <= computeQualityScore(rowNormal) {
		t.Error("favorite document should score higher")
	}
}

func TestComputeQualityScore_ValuableReview(t *testing.T) {
	rowNormal := docScoringRow{Title: "Title", WordCount: 200}
	rowValuable := docScoringRow{Title: "Title", WordCount: 200, ReviewStatus: "valuable"}
	if computeQualityScore(rowValuable) <= computeQualityScore(rowNormal) {
		t.Error("valuable review should score higher")
	}
}

func TestComputeQualityScore_MaxCapped100(t *testing.T) {
	// Max out all bonuses.
	row := docScoringRow{
		Title:          "This Is A Sufficiently Long Title With Many Words",
		WordCount:      2000,
		HeadingsJSON:   `[{"text":"h1"},{"text":"h2"},{"text":"h3"},{"text":"h4"},{"text":"h5"},{"text":"h6"},{"text":"h7"},{"text":"h8"},{"text":"h9"}]`,
		CodeBlockCount: 5,
		IsFavorite:     true,
		ReviewStatus:   "valuable",
	}
	score := computeQualityScore(row)
	if score > 100 {
		t.Errorf("score should be capped at 100, got %f", score)
	}
}

// ---------------------------------------------------------------------------
// countHeadings
// ---------------------------------------------------------------------------

func TestCountHeadings_Empty(t *testing.T) {
	got := countHeadings("")
	if got != 0 {
		t.Errorf("countHeadings empty = %d, want 0", got)
	}
}

func TestCountHeadings_ValidJSON(t *testing.T) {
	json := `[{"text":"Introduction"},{"text":"Conclusion"}]`
	got := countHeadings(json)
	if got != 2 {
		t.Errorf("countHeadings = %d, want 2", got)
	}
}

func TestCountHeadings_InvalidJSON(t *testing.T) {
	got := countHeadings("not json")
	if got != 0 {
		t.Errorf("countHeadings invalid JSON = %d, want 0", got)
	}
}

func TestCountHeadings_SingleHeading(t *testing.T) {
	json := `[{"text":"Only One"}]`
	got := countHeadings(json)
	if got != 1 {
		t.Errorf("countHeadings single = %d, want 1", got)
	}
}

// ---------------------------------------------------------------------------
// extractTagCandidates
// ---------------------------------------------------------------------------

func TestExtractTagCandidates_FromPath(t *testing.T) {
	d := docFingerprintRow{
		ID:           "doc1",
		OriginalPath: "projects/golang/notes.md",
		Title:        "",
	}
	candidates := extractTagCandidates(d)
	names := make(map[string]bool)
	for _, c := range candidates {
		names[strings.ToLower(c.name)] = true
	}
	// "projects" and "golang" should be extracted (skip stop words and 'notes' is likely stop word)
	if !names["projects"] && !names["golang"] {
		t.Errorf("expected path-derived tags, got %v", names)
	}
}

func TestExtractTagCandidates_FromTitle(t *testing.T) {
	d := docFingerprintRow{
		ID:    "doc1",
		Title: "Kubernetes Deployment Guide",
	}
	candidates := extractTagCandidates(d)
	names := make(map[string]bool)
	for _, c := range candidates {
		names[strings.ToLower(c.name)] = true
	}
	if !names["kubernetes"] && !names["deployment"] {
		t.Errorf("expected title-derived tags, got %v", names)
	}
}

func TestExtractTagCandidates_NoDuplicates(t *testing.T) {
	d := docFingerprintRow{
		ID:    "doc1",
		Title: "Golang Golang Golang",
	}
	candidates := extractTagCandidates(d)
	seen := make(map[string]int)
	for _, c := range candidates {
		seen[strings.ToLower(c.name)]++
	}
	for name, count := range seen {
		if count > 1 {
			t.Errorf("duplicate candidate %q found %d times", name, count)
		}
	}
}

func TestExtractTagCandidates_EmptyDoc(t *testing.T) {
	d := docFingerprintRow{ID: "doc1"}
	candidates := extractTagCandidates(d)
	// Should not panic, may be empty.
	_ = candidates
}

// ---------------------------------------------------------------------------
// isStopWord
// ---------------------------------------------------------------------------

func TestIsStopWord_KnownStopWords(t *testing.T) {
	stopWords := []string{"a", "an", "the", "and", "or", "but", "in", "on", "at", "to",
		"for", "of", "with", "is", "are", "it", "its", "this", "that",
		"index", "readme", "docs", "doc", "file", "files", "src", "lib",
		"untitled", "new", "note", "notes", "page", "pages", "md"}
	for _, w := range stopWords {
		if !isStopWord(w) {
			t.Errorf("isStopWord(%q) = false, want true", w)
		}
	}
}

func TestIsStopWord_NonStopWords(t *testing.T) {
	words := []string{"golang", "kubernetes", "database", "service", "microservice"}
	for _, w := range words {
		if isStopWord(w) {
			t.Errorf("isStopWord(%q) = true, want false", w)
		}
	}
}

// ---------------------------------------------------------------------------
// titleCase
// ---------------------------------------------------------------------------

func TestTitleCase_Basic(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"hello", "Hello"},
		{"WORLD", "World"},
		{"go", "Go"},
		{"", ""},
		{"x", "X"},
		{"abc", "Abc"},
	}
	for _, tc := range tests {
		got := titleCase(tc.in)
		if got != tc.want {
			t.Errorf("titleCase(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// extractHeadingText
// ---------------------------------------------------------------------------

func TestExtractHeadingText_Empty(t *testing.T) {
	got := extractHeadingText("")
	if got != "" {
		t.Errorf("extractHeadingText empty = %q, want \"\"", got)
	}
}

func TestExtractHeadingText_ValidJSON(t *testing.T) {
	json := `[{"text":"Introduction"},{"text":"Conclusion"}]`
	got := extractHeadingText(json)
	if !strings.Contains(got, "Introduction") {
		t.Errorf("extractHeadingText missing 'Introduction': %q", got)
	}
	if !strings.Contains(got, "Conclusion") {
		t.Errorf("extractHeadingText missing 'Conclusion': %q", got)
	}
}

func TestExtractHeadingText_InvalidJSON(t *testing.T) {
	got := extractHeadingText("not json")
	if got != "" {
		t.Errorf("extractHeadingText invalid JSON = %q, want \"\"", got)
	}
}

// ---------------------------------------------------------------------------
// buildDocText
// ---------------------------------------------------------------------------

func TestBuildDocText_CombinesFields(t *testing.T) {
	d := docFingerprintRow{
		Title:        "My Title",
		HeadingsJSON: `[{"text":"Section One"}]`,
		Summary:      "A brief summary.",
	}
	got := buildDocText(d)
	if !strings.Contains(got, "My Title") {
		t.Errorf("buildDocText missing title: %q", got)
	}
	if !strings.Contains(got, "Section One") {
		t.Errorf("buildDocText missing heading: %q", got)
	}
	if !strings.Contains(got, "A brief summary") {
		t.Errorf("buildDocText missing summary: %q", got)
	}
}

// ---------------------------------------------------------------------------
// min
// ---------------------------------------------------------------------------

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{0, -1, -1},
		{-3, -2, -3},
	}
	for _, tc := range tests {
		got := min(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("min(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Integration tests
// ---------------------------------------------------------------------------

func TestScoreQuality_Integration(t *testing.T) {
	svc, docSvc, _ := setupOrganize(t)
	ctx := context.Background()

	// Create documents with varying characteristics.
	_, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Short", // short title
		Content: "x",    // very short content → low quality
	})
	if err != nil {
		t.Fatalf("Create short doc: %v", err)
	}

	_, err = docSvc.Create(ctx, document.CreateRequest{
		Title:   "A Well-Structured Document With Many Words And Content",
		Content: strings.Repeat("This is a meaningful paragraph with several words in it. ", 30),
	})
	if err != nil {
		t.Fatalf("Create long doc: %v", err)
	}

	job, err := svc.ScoreQuality(ctx)
	if err != nil {
		t.Fatalf("ScoreQuality: %v", err)
	}

	if job.ProcessedItems != 2 {
		t.Errorf("ProcessedItems = %d, want 2", job.ProcessedItems)
	}
	if job.Status != JobStatusDone {
		t.Errorf("Status = %q, want %q", job.Status, JobStatusDone)
	}
}

func TestSuggestTags_Integration(t *testing.T) {
	svc, docSvc, _ := setupOrganize(t)
	ctx := context.Background()

	doc, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Kubernetes Deployment Automation Guide",
		Content: "# Kubernetes\n\nThis guide covers deployment automation strategies.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	job, err := svc.SuggestTags(ctx)
	if err != nil {
		t.Fatalf("SuggestTags: %v", err)
	}

	if job.Status != JobStatusDone {
		t.Errorf("Status = %q, want %q", job.Status, JobStatusDone)
	}

	// ProcessedItems is the total tag suggestions created.
	suggestions, err := svc.GetTagSuggestions(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetTagSuggestions: %v", err)
	}

	if len(suggestions) == 0 {
		t.Error("expected at least one tag suggestion for a descriptive document")
	}

	// Verify suggestions have required fields.
	for _, s := range suggestions {
		if s.TagName == "" {
			t.Error("suggestion TagName is empty")
		}
		if s.Source == "" {
			t.Error("suggestion Source is empty")
		}
		if s.Confidence <= 0 {
			t.Error("suggestion Confidence should be > 0")
		}
	}
}

func TestDetectSimilarity_Integration(t *testing.T) {
	svc, docSvc, _ := setupOrganize(t)
	ctx := context.Background()

	// Create two similar documents.
	content := "# Go Programming\n\n" +
		"Go is a statically typed compiled language. " +
		"It was designed at Google. " +
		"Go has garbage collection and goroutines."

	_, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Go Programming Language",
		Content: content,
	})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}

	contentSimilar := "# Go Language\n\n" +
		"Go is a statically typed compiled language. " +
		"Designed at Google with garbage collection and goroutines. " +
		"Go is very popular for backend services."

	_, err = docSvc.Create(ctx, document.CreateRequest{
		Title:   "Go Language Overview",
		Content: contentSimilar,
	})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}

	// Create a clearly different document.
	_, err = docSvc.Create(ctx, document.CreateRequest{
		Title:   "Cooking Pasta Recipes",
		Content: "# Pasta\n\nBoil water add pasta stir occasionally drain serve with sauce.",
	})
	if err != nil {
		t.Fatalf("Create C: %v", err)
	}

	job, err := svc.DetectSimilarity(ctx)
	if err != nil {
		t.Fatalf("DetectSimilarity: %v", err)
	}

	if job.Status != JobStatusDone {
		t.Errorf("Status = %q, want %q", job.Status, JobStatusDone)
	}
	// At minimum the job should complete without error; the similar Go docs
	// may or may not exceed the threshold depending on content overlap.
}

func TestDetectPromptCandidatesFTSBased_Integration(t *testing.T) {
	svc, docSvc, _ := setupOrganize(t)
	ctx := context.Background()

	// Create a document that looks like a prompt.
	_, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "System Prompt Template",
		Content: "# System Prompt\n\nYou are an expert assistant. prompt user: assistant: Please help.",
	})
	if err != nil {
		t.Fatalf("Create prompt doc: %v", err)
	}

	// Create a document that is not a prompt.
	_, err = docSvc.Create(ctx, document.CreateRequest{
		Title:   "Regular Meeting Notes",
		Content: "# Meeting Notes\n\nDiscussed Q3 budget and project milestones.",
	})
	if err != nil {
		t.Fatalf("Create normal doc: %v", err)
	}

	job, err := svc.DetectPromptCandidates(ctx)
	if err != nil {
		t.Fatalf("DetectPromptCandidates: %v", err)
	}

	if job.Status != JobStatusDone {
		t.Errorf("Status = %q, want done", job.Status)
	}
}

func TestAcceptTagSuggestion_Integration(t *testing.T) {
	svc, docSvc, _ := setupOrganize(t)
	ctx := context.Background()

	doc, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Docker Container Orchestration",
		Content: "# Docker\n\nContainer orchestration with Docker Compose.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	job, err := svc.SuggestTags(ctx)
	if err != nil {
		t.Fatalf("SuggestTags: %v", err)
	}
	if job.Status != JobStatusDone {
		t.Fatalf("SuggestTags status = %q", job.Status)
	}

	suggestions, err := svc.GetTagSuggestions(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetTagSuggestions: %v", err)
	}
	if len(suggestions) == 0 {
		t.Skip("no suggestions generated, skipping accept test")
	}

	// Accept the first suggestion.
	err = svc.AcceptTagSuggestion(ctx, suggestions[0].ID)
	if err != nil {
		t.Fatalf("AcceptTagSuggestion: %v", err)
	}

	// Verify suggestion is now accepted.
	remaining, err := svc.GetTagSuggestions(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetTagSuggestions after accept: %v", err)
	}
	for _, s := range remaining {
		if s.ID == suggestions[0].ID && s.Status != TagSuggestionStatusAccepted {
			t.Errorf("suggestion status = %q, want %q", s.Status, TagSuggestionStatusAccepted)
		}
	}
}

func TestRejectTagSuggestion_Integration(t *testing.T) {
	svc, docSvc, _ := setupOrganize(t)
	ctx := context.Background()

	doc, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Python Machine Learning Tutorial",
		Content: "# Python ML\n\nMachine learning with Python scikit-learn.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = svc.SuggestTags(ctx)
	if err != nil {
		t.Fatalf("SuggestTags: %v", err)
	}

	suggestions, err := svc.GetTagSuggestions(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetTagSuggestions: %v", err)
	}
	if len(suggestions) == 0 {
		t.Skip("no suggestions generated, skipping reject test")
	}

	err = svc.RejectTagSuggestion(ctx, suggestions[0].ID)
	if err != nil {
		t.Fatalf("RejectTagSuggestion: %v", err)
	}
}

func TestGetSimilarDocuments_Integration(t *testing.T) {
	svc, docSvc, _ := setupOrganize(t)
	ctx := context.Background()

	doc, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Sample Document",
		Content: "# Sample\n\nSome content.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	similar, err := svc.GetSimilarDocuments(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetSimilarDocuments: %v", err)
	}
	// Empty result is fine, just should not error.
	_ = similar
}

func TestListJobs_Integration(t *testing.T) {
	svc, docSvc, _ := setupOrganize(t)
	ctx := context.Background()

	_, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Test Document",
		Content: "Content for testing.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = svc.ScoreQuality(ctx)
	if err != nil {
		t.Fatalf("ScoreQuality: %v", err)
	}

	jobs, err := svc.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) == 0 {
		t.Error("expected at least one job in list")
	}
}
