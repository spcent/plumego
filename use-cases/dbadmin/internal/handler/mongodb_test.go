package handler

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- validateName ---

func TestValidateName_valid(t *testing.T) {
	valid := []string{
		"users",
		"my_collection",
		"data2024",
		"test-db",
	}
	for _, name := range valid {
		if err := validateName(name); err != nil {
			t.Errorf("validateName(%q) error=%v, want nil", name, err)
		}
	}
}

func TestValidateName_invalid(t *testing.T) {
	cases := []struct {
		name   string
		reason string
	}{
		{"", "empty name"},
		{"test.db", "contains dot"},
		{"test db", "contains space"},
		{"test$db", "contains dollar"},
		{"test/db", "contains slash"},
		{"test\\db", "contains backslash"},
		{string(rune(0)), "contains null"},
	}
	for _, c := range cases {
		if err := validateName(c.name); err == nil {
			t.Errorf("validateName(%q) should error for %s", c.name, c.reason)
		}
	}
}

func TestValidateName_tooLong(t *testing.T) {
	long := ""
	for i := 0; i < 65; i++ {
		long += "a"
	}
	if err := validateName(long); err == nil {
		t.Errorf("validateName should error for name > 64 chars")
	}
}

// --- parseID ---

func TestParseID_objectId(t *testing.T) {
	// Valid 24-char hex ObjectId
	idStr := "507f1f77bcf86cd799439011"
	id, err := parseID(idStr)
	if err != nil {
		t.Fatalf("parseID(%q) error=%v", idStr, err)
	}
	oid, ok := id.(primitive.ObjectID)
	if !ok {
		t.Fatalf("parseID(%q) type=%T, want primitive.ObjectID", idStr, id)
	}
	if oid.Hex() != idStr {
		t.Errorf("parseID(%q) hex=%q, want %q", idStr, oid.Hex(), idStr)
	}
}

func TestParseID_number(t *testing.T) {
	id, err := parseID("123")
	if err != nil {
		t.Fatalf("parseID(123) error=%v", err)
	}
	// parseID accepts only ObjectID hex strings or plain string IDs.
	// This prevents user-controlled JSON values from being interpreted as other BSON scalar types.
	// So "123" is returned as string "123", not float64 123.0.
	if s, ok := id.(string); !ok || s != "123" {
		t.Errorf("parseID(123)=%v (%T), want \"123\" (string)", id, id)
	}
}

func TestParseID_string(t *testing.T) {
	id, err := parseID("my-custom-id")
	if err != nil {
		t.Fatalf("parseID(my-custom-id) error=%v", err)
	}
	if s, ok := id.(string); !ok || s != "my-custom-id" {
		t.Errorf("parseID(my-custom-id)=%v (%T), want my-custom-id (string)", id, id)
	}
}

// --- convertObjectIDs ---

func TestConvertObjectIDs_topLevel(t *testing.T) {
	doc := bson.M{
		"_id": map[string]any{
			"$oid": "507f1f77bcf86cd799439011",
		},
		"name": "test",
	}
	convertObjectIDs(doc)
	oid, ok := doc["_id"].(primitive.ObjectID)
	if !ok {
		t.Fatalf("convertObjectIDs did not convert _id to ObjectID, got %T", doc["_id"])
	}
	if oid.Hex() != "507f1f77bcf86cd799439011" {
		t.Errorf("convertObjectIDs hex=%q, want 507f1f77bcf86cd799439011", oid.Hex())
	}
}

func TestConvertObjectIDs_nested(t *testing.T) {
	doc := bson.M{
		"user": map[string]any{
			"id": map[string]any{
				"$oid": "507f1f77bcf86cd799439011",
			},
		},
	}
	convertObjectIDs(doc)
	// convertObjectIDs recursively converts nested maps
	user, ok := doc["user"].(map[string]any)
	if !ok {
		t.Fatalf("user is not map[string]any, got %T", doc["user"])
	}
	oid, ok := user["id"].(primitive.ObjectID)
	if !ok {
		t.Fatalf("convertObjectIDs did not convert nested id, got %T", user["id"])
	}
	if oid.Hex() != "507f1f77bcf86cd799439011" {
		t.Errorf("convertObjectIDs hex=%q, want 507f1f77bcf86cd799439011", oid.Hex())
	}
}

func TestConvertObjectIDs_array(t *testing.T) {
	doc := bson.M{
		"refs": []any{
			map[string]any{
				"$oid": "507f1f77bcf86cd799439011",
			},
		},
	}
	convertObjectIDs(doc)
	refs := doc["refs"].([]any)
	oid, ok := refs[0].(primitive.ObjectID)
	if !ok {
		t.Fatalf("convertObjectIDs did not convert array element, got %T", refs[0])
	}
	if oid.Hex() != "507f1f77bcf86cd799439011" {
		t.Errorf("convertObjectIDs hex=%q, want 507f1f77bcf86cd799439011", oid.Hex())
	}
}

// --- bsonToJSON ---

func TestBsonToJSON_objectId(t *testing.T) {
	oid, _ := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	result := bsonToJSON(oid)
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("bsonToJSON(ObjectID) type=%T, want map[string]any", result)
	}
	if m["$oid"] != "507f1f77bcf86cd799439011" {
		t.Errorf("bsonToJSON(ObjectID)=$oid=%v, want 507f1f77bcf86cd799439011", m["$oid"])
	}
}

func TestBsonToJSON_string(t *testing.T) {
	result := bsonToJSON("test")
	if s, ok := result.(string); !ok || s != "test" {
		t.Errorf("bsonToJSON(test)=%v (%T), want test (string)", result, result)
	}
}

func TestBsonToJSON_int(t *testing.T) {
	result := bsonToJSON(42)
	if n, ok := result.(int); !ok || n != 42 {
		t.Errorf("bsonToJSON(42)=%v (%T), want 42 (int)", result, result)
	}
}

// --- dangerous aggregation detection ---

func TestDetectDangerousAggregation_out(t *testing.T) {
	pipeline := []any{
		map[string]any{"$match": map[string]any{"status": "active"}},
		map[string]any{"$out": "output_collection"},
	}
	hasDanger := false
	for _, stage := range pipeline {
		if stageMap, ok := stage.(map[string]any); ok {
			if _, hasOut := stageMap["$out"]; hasOut {
				hasDanger = true
				break
			}
		}
	}
	if !hasDanger {
		t.Errorf("detectDangerousAggregation should detect $out")
	}
}

func TestDetectDangerousAggregation_merge(t *testing.T) {
	pipeline := []any{
		map[string]any{"$match": map[string]any{"status": "active"}},
		map[string]any{"$merge": map[string]any{"into": "output"}},
	}
	hasDanger := false
	for _, stage := range pipeline {
		if stageMap, ok := stage.(map[string]any); ok {
			if _, hasMerge := stageMap["$merge"]; hasMerge {
				hasDanger = true
				break
			}
		}
	}
	if !hasDanger {
		t.Errorf("detectDangerousAggregation should detect $merge")
	}
}

func TestDetectDangerousAggregation_safe(t *testing.T) {
	pipeline := []any{
		map[string]any{"$match": map[string]any{"status": "active"}},
		map[string]any{"$group": map[string]any{"_id": "$category"}},
	}
	hasDanger := false
	for _, stage := range pipeline {
		if stageMap, ok := stage.(map[string]any); ok {
			if _, hasOut := stageMap["$out"]; hasOut {
				hasDanger = true
				break
			}
			if _, hasMerge := stageMap["$merge"]; hasMerge {
				hasDanger = true
				break
			}
		}
	}
	if hasDanger {
		t.Errorf("detectDangerousAggregation should not flag safe pipeline")
	}
}

// --- JSON filter parsing ---

func TestParseJSONFilter_valid(t *testing.T) {
	filterStr := `{"status": "active", "age": {"$gt": 18}}`
	filter, err := parseMongoFilterJSON(filterStr)
	if err != nil {
		t.Fatalf("parseMongoFilterJSON error=%v", err)
	}
	if filter["status"] != "active" {
		t.Errorf("filter[status]=%v, want active", filter["status"])
	}
}

func TestParseJSONFilter_invalid(t *testing.T) {
	filterStr := `{invalid json`
	if _, err := parseMongoFilterJSON(filterStr); err == nil {
		t.Errorf("parseMongoFilterJSON should error for invalid JSON")
	}
}

func TestParseMongoFilterJSON_rejectsServerSideCode(t *testing.T) {
	cases := []string{
		`{"$where":"this.secret == true"}`,
		`{"status":{"$function":{"body":"function(){return true}","args":[],"lang":"js"}}}`,
		`{"$expr":{"$accumulator":{"init":"function(){}"}}}`,
		`{"items":{"$elemMatch":{"$expr":{"$gt":["$a","$b"]}}}}`,
	}
	for _, raw := range cases {
		if _, err := parseMongoFilterJSON(raw); err == nil {
			t.Errorf("parseMongoFilterJSON(%s) = nil, want error", raw)
		}
	}
}

func TestParseMongoProjectionJSON_rejectsOperators(t *testing.T) {
	if _, err := parseMongoProjectionJSON(`{"name":1,"_id":0}`); err != nil {
		t.Fatalf("valid projection rejected: %v", err)
	}
	if _, err := parseMongoProjectionJSON(`{"$where":1}`); err == nil {
		t.Fatal("expected projection operator to be rejected")
	}
}

func TestParseMongoSortJSON_rejectsInvalidDirection(t *testing.T) {
	if _, err := parseMongoSortJSON(`{"created_at":-1,"name":"asc"}`); err != nil {
		t.Fatalf("valid sort rejected: %v", err)
	}
	if _, err := parseMongoSortJSON(`{"name":2}`); err == nil {
		t.Fatal("expected invalid sort direction to be rejected")
	}
}

func TestParseMongoPipelineJSON_rejectsUnsupportedStage(t *testing.T) {
	if _, _, err := parseMongoPipelineJSON(`[{"$match":{"status":"active"}},{"$limit":10}]`); err != nil {
		t.Fatalf("valid pipeline rejected: %v", err)
	}
	if _, _, err := parseMongoPipelineJSON(`[{"$where":"this.secret"}]`); err == nil {
		t.Fatal("expected unsupported pipeline stage to be rejected")
	}
}

func TestParseMongoFilterJSON_rejectsDeeplyNested(t *testing.T) {
	// Build a document nested one level beyond maxMongoFilterDepth.
	root := map[string]any{}
	cur := root
	for i := 0; i < maxMongoFilterDepth+1; i++ {
		child := map[string]any{}
		cur["field"] = child
		cur = child
	}
	if err := validateMongoFilterDocument(root); err == nil {
		t.Error("validateMongoFilterDocument should reject document nested beyond maxMongoFilterDepth")
	}
}

func TestParseMongoPipelineJSON_detectsDangerousWriteStage(t *testing.T) {
	_, dangerous, err := parseMongoPipelineJSON(`[{"$match":{"status":"active"}},{"$merge":{"into":"archive"}}]`)
	if err != nil {
		t.Fatalf("parseMongoPipelineJSON returned error: %v", err)
	}
	if !dangerous {
		t.Fatal("expected $merge to be detected as dangerous")
	}
}

// --- getSortedKeys ---

func TestGetSortedKeys(t *testing.T) {
	doc := bson.M{
		"zebra": 1,
		"apple": 2,
		"mango": 3,
	}
	keys := getSortedKeys(doc)
	if len(keys) != 3 {
		t.Fatalf("getSortedKeys len=%d, want 3", len(keys))
	}
	if keys[0] != "apple" || keys[1] != "mango" || keys[2] != "zebra" {
		t.Errorf("getSortedKeys=%v, want [apple mango zebra]", keys)
	}
}
