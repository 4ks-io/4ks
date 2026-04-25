package recipesvc

import "testing"

func TestCreateMockBanner(t *testing.T) {
	t.Parallel()

	service := recipeService{}
	banner := service.CreateMockBanner("fallback.jpg", "https://cdn.example/fallback.jpg")

	if len(banner) != 2 {
		t.Fatalf("expected two banner variants, got %d", len(banner))
	}
	if banner[0].Alias != "sm" || banner[0].MaxWidth != 256 {
		t.Fatalf("unexpected small banner variant: %+v", banner[0])
	}
	if banner[1].Alias != "md" || banner[1].MaxWidth != 1024 {
		t.Fatalf("unexpected medium banner variant: %+v", banner[1])
	}
	if banner[0].URL != "https://cdn.example/fallback.jpg" || banner[1].Filename != "fallback.jpg" {
		t.Fatalf("expected filename and URL to be propagated, got %+v", banner)
	}
}
