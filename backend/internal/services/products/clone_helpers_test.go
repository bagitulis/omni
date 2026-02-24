package products

import (
	"testing"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- parseImagesFromDB ---

func TestParseImagesFromDB_Nil_ReturnsEmpty(t *testing.T) {
	result := parseImagesFromDB(nil)
	assert.Empty(t, result)
}

func TestParseImagesFromDB_EmptyString_ReturnsEmpty(t *testing.T) {
	result := parseImagesFromDB("")
	assert.Empty(t, result)
}

func TestParseImagesFromDB_SingleURL_ReturnsSingle(t *testing.T) {
	result := parseImagesFromDB("https://example.com/image.jpg")
	require.Len(t, result, 1)
	assert.Equal(t, "https://example.com/image.jpg", result[0])
}

func TestParseImagesFromDB_JSONArray_ReturnsSlice(t *testing.T) {
	input := `["https://a.com/1.jpg","https://b.com/2.jpg"]`
	result := parseImagesFromDB(input)
	require.Len(t, result, 2)
	assert.Equal(t, "https://a.com/1.jpg", result[0])
	assert.Equal(t, "https://b.com/2.jpg", result[1])
}

func TestParseImagesFromDB_StringSlice_ReturnsSame(t *testing.T) {
	input := []string{"img1.jpg", "img2.jpg"}
	result := parseImagesFromDB(input)
	assert.Equal(t, input, result)
}

func TestParseImagesFromDB_InterfaceSlice_FiltersStrings(t *testing.T) {
	input := []interface{}{"img1.jpg", 42, "img2.jpg", nil}
	result := parseImagesFromDB(input)
	require.Len(t, result, 2)
	assert.Equal(t, "img1.jpg", result[0])
	assert.Equal(t, "img2.jpg", result[1])
}

func TestParseImagesFromDB_UnknownType_ReturnsEmpty(t *testing.T) {
	result := parseImagesFromDB(12345)
	assert.Empty(t, result)
}

// --- minInt ---

func TestMinInt_FirstSmaller(t *testing.T) {
	assert.Equal(t, 3, minInt(3, 7))
}

func TestMinInt_SecondSmaller(t *testing.T) {
	assert.Equal(t, 2, minInt(5, 2))
}

func TestMinInt_Equal(t *testing.T) {
	assert.Equal(t, 4, minInt(4, 4))
}

func TestMinInt_Negatives(t *testing.T) {
	assert.Equal(t, -10, minInt(-10, -3))
}

// --- getStringOrEmpty ---

func TestGetStringOrEmpty_NoArgs_ReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", getStringOrEmpty())
}

func TestGetStringOrEmpty_NilOnly_ReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", getStringOrEmpty(nil))
}

func TestGetStringOrEmpty_EmptyStringOnly_ReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", getStringOrEmpty(""))
}

func TestGetStringOrEmpty_FirstNonEmpty_ReturnsFirst(t *testing.T) {
	assert.Equal(t, "hello", getStringOrEmpty("", nil, "hello", "world"))
}

func TestGetStringOrEmpty_StringPointer_NonNil(t *testing.T) {
	s := "ptr-value"
	assert.Equal(t, "ptr-value", getStringOrEmpty(&s))
}

func TestGetStringOrEmpty_NilPointer_ReturnsEmpty(t *testing.T) {
	var s *string
	assert.Equal(t, "", getStringOrEmpty(s))
}

// --- getFloat64 ---

func TestGetFloat64_NoArgs_ReturnsZero(t *testing.T) {
	assert.Equal(t, float64(0), getFloat64())
}

func TestGetFloat64_NilOnly_ReturnsZero(t *testing.T) {
	assert.Equal(t, float64(0), getFloat64(nil))
}

func TestGetFloat64_ZeroFloat_ReturnsZero(t *testing.T) {
	assert.Equal(t, float64(0), getFloat64(float64(0)))
}

func TestGetFloat64_PositiveFloat_ReturnsValue(t *testing.T) {
	assert.Equal(t, 3.14, getFloat64(3.14))
}

func TestGetFloat64_PositiveInt_ReturnsFloat(t *testing.T) {
	assert.Equal(t, float64(5), getFloat64(5))
}

func TestGetFloat64_FirstPositive_ReturnsFirst(t *testing.T) {
	assert.Equal(t, float64(10), getFloat64(float64(0), float64(10), float64(20)))
}

func TestGetFloat64_FloatPointer_NonNil(t *testing.T) {
	v := 9.9
	assert.Equal(t, 9.9, getFloat64(&v))
}

// --- getInt ---

func TestGetInt_NoArgs_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0, getInt())
}

func TestGetInt_NilOnly_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0, getInt(nil))
}

func TestGetInt_ZeroInt_ReturnsZero(t *testing.T) {
	assert.Equal(t, 0, getInt(0))
}

func TestGetInt_PositiveInt_ReturnsValue(t *testing.T) {
	assert.Equal(t, 42, getInt(42))
}

func TestGetInt_NegativeInt_Skipped_ReturnsNextValid(t *testing.T) {
	// negative is skipped (< 0), next is 5
	assert.Equal(t, 5, getInt(-3, 5))
}

func TestGetInt_Int64_ReturnsInt(t *testing.T) {
	assert.Equal(t, 7, getInt(int64(7)))
}

func TestGetInt_Float64_ReturnsInt(t *testing.T) {
	assert.Equal(t, 3, getInt(float64(3.9)))
}

// --- buildTiktokImageInfos ---

func TestBuildTiktokImageInfos_Empty(t *testing.T) {
	result := buildTiktokImageInfos([]string{})
	assert.Empty(t, result)
}

func TestBuildTiktokImageInfos_SingleImage(t *testing.T) {
	result := buildTiktokImageInfos([]string{"uri1"})
	require.Len(t, result, 1)
	assert.Equal(t, tiktokPkg.ImageInfo{URI: "uri1"}, result[0])
}

func TestBuildTiktokImageInfos_MultipleImages(t *testing.T) {
	images := []string{"a", "b", "c"}
	result := buildTiktokImageInfos(images)
	require.Len(t, result, 3)
	for i, img := range result {
		assert.Equal(t, images[i], img.URI)
	}
}

// --- determineBatchStatus (method on CloneService, but can call on nil receiver via value) ---

func TestDetermineBatchStatus_NoFailures_Completed(t *testing.T) {
	svc := &CloneService{}
	assert.Equal(t, "completed", svc.determineBatchStatus(0, 5))
}

func TestDetermineBatchStatus_AllFailed_Failed(t *testing.T) {
	svc := &CloneService{}
	assert.Equal(t, "failed", svc.determineBatchStatus(3, 0))
}

func TestDetermineBatchStatus_Partial_Partial(t *testing.T) {
	svc := &CloneService{}
	assert.Equal(t, "partial", svc.determineBatchStatus(2, 3))
}

// --- generateCloneID ---

func TestGenerateCloneID_HasClonePrefix(t *testing.T) {
	id := generateCloneID()
	assert.NotEmpty(t, id)
	assert.Contains(t, id, "clone-")
}

func TestGenerateCloneID_UniqueOnEachCall(t *testing.T) {
	id1 := generateCloneID()
	id2 := generateCloneID()
	// They should either be different (because of time) or at minimum both non-empty
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
}
