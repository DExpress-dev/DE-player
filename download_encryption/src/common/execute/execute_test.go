package execute

import (
	"fmt"
	"testing"
)

func TestDown(t *testing.T) {
	url := "https://www.videoindexer.ai/Api/Widget/Breakdowns/6b540ebafe/6b540ebafe/Vtt?language=Chinese&accessToken=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJFeHRlcm5hbFVzZXJJZCI6ImNiNjU2ZjM4OTU5YWM5ZWMiLCJVc2VyVHlwZSI6Ik1pY3Jvc29mdCIsIkJyZWFrZG93bklkIjoiNmI1NDBlYmFmZSIsIkFsbG93RWRpdCI6IkZhbHNlIiwiaXNzIjoiaHR0cHM6Ly93d3cudmlkZW9pbmRleGVyLmFpIiwiYXVkIjoiaHR0cHM6Ly93d3cudmlkZW9pbmRleGVyLmFpIiwiZXhwIjoxNTE3OTAyOTI4LCJuYmYiOjE1MTc4OTkwMjh9.i6_R38mkhrTVX_H-d7Uz_6_q08UjJMexF7Ut2QJ_FDI"
	fmt.Println(DoLoadFile(url))
}
