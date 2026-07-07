package identity

var DesktopUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0",
}

var MobileUserAgents = []string{
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 14; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/120.0.6099.119 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (iPad; CPU OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 13; SM-X906C) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

var AcceptLanguages = []string{
	"en-US,en;q=0.9",
	"en-GB,en;q=0.9",
	"da-DK,da;q=0.9,en-US;q=0.8,en;q=0.7",
	"fr-FR,fr;q=0.9,en-US;q=0.8",
}

var Referers = []string{
	"https://facebook.com/",
	"https://twitter.com/",
	"https://linkedin.com/",
	"https://google.com/",
	"https://reddit.com/",
	"https://youtube.com/",
	"https://discord.com/",
	"https://slack.com/",
	"https://whatsapp.com/",
	"https://tiktok.com/",
	"https://pinterest.com/",
	"https://telegram.org/",
	"https://weibo.com/",
}

var Locations = []Location{
	{Country: "US", City: "The%20Dalles", Region: "OR", Latitude: "45.5946", Longitude: "-121.1787"},
	{Country: "US", City: "Atlanta", Region: "GA", Latitude: "33.7490", Longitude: "-84.3880"},
	{Country: "US", City: "New%20York", Region: "NY", Latitude: "40.7128", Longitude: "-74.0060"},
	{Country: "US", City: "San%20Francisco", Region: "CA", Latitude: "37.7749", Longitude: "-122.4194"},
	{Country: "DK", City: "Copenhagen", Region: "84", Latitude: "55.6761", Longitude: "12.5683"},
	{Country: "DK", City: "Aarhus", Region: "82", Latitude: "56.1629", Longitude: "10.2039"},
	{Country: "DE", City: "Munich", Region: "BY", Latitude: "48.1351", Longitude: "11.5820"},
	{Country: "GB", City: "London", Region: "ENG", Latitude: "51.5074", Longitude: "-0.1278"},
	{Country: "FR", City: "Paris", Region: "IDF", Latitude: "48.8566", Longitude: "2.3522"},
	{Country: "NL", City: "Amsterdam", Region: "NH", Latitude: "52.3676", Longitude: "4.9041"},
	{Country: "SE", City: "Stockholm", Region: "AB", Latitude: "59.3293", Longitude: "18.0686"},
	{Country: "JP", City: "Tokyo", Region: "13", Latitude: "35.6762", Longitude: "139.6503"},
	{Country: "BR", City: "S%C3%A3o%20Paulo", Region: "SP", Latitude: "-23.5505", Longitude: "-46.6333"},
	{Country: "AU", City: "Sydney", Region: "NSW", Latitude: "-33.8688", Longitude: "151.2093"},
}

var IPFirstOctets = map[string][]int{
	"US": []int{3, 8, 13, 18, 23, 34, 35, 44, 50, 52, 54, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 96, 97, 98, 99, 100, 104, 107, 108, 128, 129, 130, 131, 132, 134, 135, 136, 137, 138, 139, 140, 142, 143, 144, 147, 148, 149, 150, 151, 152, 153, 154, 155, 156, 157, 158, 159, 160, 161, 162, 163, 164, 165, 166, 167, 168, 170, 171, 172, 173, 174, 184, 198, 199, 204, 205, 206, 207, 208, 209},
	"DK": []int{2, 5, 31, 37, 46, 77, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 109, 176, 178, 185, 188, 193, 194, 195, 212, 213},
	"DE": []int{2, 5, 31, 37, 46, 62, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 91, 92, 93, 94, 95, 109, 134, 138, 141, 145, 146, 176, 178, 185, 188, 193, 194, 195, 212, 213, 217},
	"GB": []int{2, 5, 31, 37, 46, 51, 62, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 109, 176, 178, 185, 188, 193, 194, 195, 212, 213, 217},
	"FR": []int{2, 5, 31, 37, 46, 62, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 109, 176, 178, 185, 188, 193, 194, 195, 212, 213, 217},
	"NL": []int{2, 5, 31, 37, 46, 62, 77, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 91, 92, 93, 94, 95, 109, 145, 176, 178, 185, 188, 193, 194, 195, 212, 213},
	"SE": []int{2, 5, 31, 37, 46, 62, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 91, 92, 93, 94, 95, 109, 176, 178, 185, 188, 193, 194, 195, 212, 213},
	"JP": []int{1, 14, 27, 36, 42, 49, 58, 59, 60, 61, 101, 106, 110, 111, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 133, 150, 153, 157, 160, 163, 175, 180, 182, 183, 202, 210, 211, 218, 219, 220},
	"BR": []int{131, 138, 139, 143, 146, 152, 155, 161, 168, 170, 177, 179, 186, 187, 189, 190, 191, 200, 201},
	"AU": []int{1, 14, 27, 43, 49, 58, 59, 60, 61, 101, 103, 106, 110, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 144, 150, 175, 180, 182, 192, 202, 203, 210, 211, 218, 219, 220},
}
