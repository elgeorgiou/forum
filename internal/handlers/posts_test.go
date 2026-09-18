package handlers

import (
	"reflect"
	"testing"
)

func TestParseCategoryIDs(t *testing.T) {
	categoryIDs, err := parseCategoryIDs(
		[]string{"1", "2", "3"},
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	expected := []int64{1, 2, 3}

	if !reflect.DeepEqual(
		categoryIDs,
		expected,
	) {
		t.Fatalf(
			"expected %v, got %v",
			expected,
			categoryIDs,
		)
	}
}

func TestParseCategoryIDsRemovesDuplicates(
	t *testing.T,
) {
	categoryIDs, err := parseCategoryIDs(
		[]string{"1", "2", "1", "2"},
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	expected := []int64{1, 2}

	if !reflect.DeepEqual(
		categoryIDs,
		expected,
	) {
		t.Fatalf(
			"expected %v, got %v",
			expected,
			categoryIDs,
		)
	}
}

func TestParseCategoryIDsRejectsInvalidID(
	t *testing.T,
) {
	_, err := parseCategoryIDs(
		[]string{"abc"},
	)

	if err == nil {
		t.Fatal(
			"expected invalid category id error",
		)
	}
}

func TestParseCategoryIDsRejectsZero(
	t *testing.T,
) {
	_, err := parseCategoryIDs(
		[]string{"0"},
	)

	if err == nil {
		t.Fatal(
			"expected zero category id to fail",
		)
	}
}

func TestParseCategoryIDsRejectsNegativeID(
	t *testing.T,
) {
	_, err := parseCategoryIDs(
		[]string{"-1"},
	)

	if err == nil {
		t.Fatal(
			"expected negative category id to fail",
		)
	}
}

func TestParseCategoryIDsSkipsEmptyValues(
	t *testing.T,
) {
	categoryIDs, err := parseCategoryIDs(
		[]string{"", " ", "2"},
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	expected := []int64{2}

	if !reflect.DeepEqual(
		categoryIDs,
		expected,
	) {
		t.Fatalf(
			"expected %v, got %v",
			expected,
			categoryIDs,
		)
	}
}
