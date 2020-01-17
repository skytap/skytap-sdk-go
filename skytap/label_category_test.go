package skytap

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestCreateLabelCategoryList(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.RequestURI, "/v2/label_categories?count=100&offset=0")
		assert.Equal(t, req.Method, "GET")
		_, err := io.WriteString(rw, `[{"id": "12345", "name": "label-test", "single_value": "true", "enabled":  true}]`)
		assert.NoError(t, err)
	}

	labelCategories, err := skytap.LabelCategory.List(context.Background())
	assert.Nil(t, err)
	assert.NotNil(t, labelCategories)
	assert.Len(t, labelCategories, 1)
	assert.Equal(t, *labelCategories[0].ID, 12345)
}

func TestCreateLabelCategoryNeverCreated(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var created = false

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err)
		assert.Equal(t, req.RequestURI, "/v2/label_categories")
		assert.Equal(t, req.Method, "POST")
		assert.JSONEq(t, `{"name":"label-test","single_value": true}`, string(body))
		_, err = io.WriteString(rw, `{"id": "12345", "name": "label-test", "single_value": "true", "enabled":  true}`)
		assert.NoError(t, err)

		created = true
	}

	opts := LabelCategory{
		Name:        strToPtr("label-test"),
		SingleValue: boolToPtr(true),
	}

	labelCategory, err := skytap.LabelCategory.Create(context.Background(), &opts)

	assert.Nil(t, err)
	assert.NotNil(t, labelCategory)
	assert.True(t, created)
	assert.Equal(t, *labelCategory.ID, 12345)
}

func TestCreateLabelCategoryWhenDisabled(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	var tryCreate = false
	var created = false

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.RequestURI == "/v2/label_categories" && req.Method == "POST" {
			tryCreate = true
			rw.WriteHeader(409)
			_, err := io.WriteString(rw, `{"error": "Category already exists. Please re-enable the existing category instead.", 
											  "url": "https://cloud.skytap.com/v2/label_categories/12345"}`)
			assert.Nil(t, err)
		}
		if req.RequestURI == "/v2/label_categories/12345.json" && req.Method == "PUT" {
			created = true
			body, err := ioutil.ReadAll(req.Body)
			assert.Nil(t, err)
			assert.JSONEq(t, `{"enabled": true}`, string(body))
			_, err = io.WriteString(rw, `{"name": "label-test", "single_value": true, "enabled":  true}`)
			assert.NoError(t, err)
		}
	}

	opts := LabelCategory{
		Name:        strToPtr("label-test"),
		SingleValue: boolToPtr(true),
	}

	labelCategory, err := skytap.LabelCategory.Create(context.Background(), &opts)

	assert.Nil(t, err)
	assert.NotNil(t, labelCategory)
	assert.True(t, tryCreate)
	assert.True(t, created)
	assert.Equal(t, *labelCategory.ID, 12345)
}

func TestCreateLabelCategoryWithConflict(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		if req.RequestURI == "/v2/label_categories" && req.Method == "POST" {
			rw.WriteHeader(409)
			_, err := io.WriteString(rw, `{"error": "Category already exists. Please re-enable the existing category instead.", 
											  "url": "https://cloud.skytap.com/v2/label_categories/12345"}`)
			assert.Nil(t, err)
		}
		if req.RequestURI == "/v2/label_categories/12345.json" && req.Method == "PUT" {
			_, err := io.WriteString(rw, `{"name": "label-test", "single_value": false, "enabled":  true}`)
			assert.NoError(t, err)
		}
	}

	opts := LabelCategory{
		Name:        strToPtr("label-test"),
		SingleValue: boolToPtr(true),
	}

	labelCategory, err := skytap.LabelCategory.Create(context.Background(), &opts)

	assert.NotNil(t, err)
	assert.Nil(t, labelCategory)
	assert.Error(t, err)
	assert.EqualError(t, err, "The label category with id: 12345 can not be created with this single value property as"+
		" it is recreated from a existing label category.")
}

func TestGetLabelCategory(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	read := false
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		read = true
		assert.Equal(t, req.Method, "GET")
		assert.Equal(t, req.RequestURI, "/v2/label_categories/12345")
		_, err := io.WriteString(rw, `{"id": "12345", "name": "label-test", "single_value": true, "enabled":  true}`)
		assert.NoError(t, err)
	}

	labelCategory, err := skytap.LabelCategory.Get(context.Background(), 12345)

	assert.Nil(t, err)
	assert.NotNil(t, labelCategory)
	assert.True(t, read)

	assert.Equal(t, *labelCategory.ID, 12345)
	assert.Equal(t, *labelCategory.SingleValue, true)
}

func TestLabelCategoryDelete(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	deleted := false
	*handler = func(rw http.ResponseWriter, req *http.Request) {
		deleted = true
		assert.Equal(t, req.RequestURI, "/v2/label_categories/12345.json")
		assert.Equal(t, req.Method, "PUT")
		body, err := ioutil.ReadAll(req.Body)
		assert.Nil(t, err)
		assert.JSONEq(t, `{"enabled": false}`, string(body))
	}

	err := skytap.LabelCategory.Delete(context.Background(), 12345)

	assert.Nil(t, err)
	assert.True(t, deleted)
}

func TestLabelCategoryWithValidationError(t *testing.T) {
	skytap, hs, handler := createClient(t)
	defer hs.Close()

	*handler = func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.RequestURI, "/v2/label_categories")
		assert.Equal(t, req.Method, "POST")
		rw.WriteHeader(409)
		_, err := io.WriteString(rw, `{"error": "Validation failed: The label category name already exists.", 
											  "url": "https://cloud.skytap.com/v2/label_categories/12345"}`)
		assert.Nil(t, err)
	}

	opts := LabelCategory{
		Name:        strToPtr("label-test"),
		SingleValue: boolToPtr(true),
	}
	_, err := skytap.LabelCategory.Create(context.Background(), &opts)

	assert.Error(t, err)
	assert.EqualError(t, err, "Error creating label category with name (label-test): Validation failed: "+
		"The label category name already exists.")
}
