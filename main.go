package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"unicode"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

var (
	TILE_SIZE       = 126
	SMALL_TILE_SIZE = 42
)

var WIND_OFFSETS = [...]int{10, 20, 30, 40, 50, 100, 150, 200, 250}

func init() {
	flag.Parse()
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "arty",
			Description: "calculate artillery",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "from",
					Description: "Coords to shoot from.  In the form of X-Y-Numpad(s).  For example:  b-2-3-4",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "to",
					Description: "Coords to shoot to.  In the form of X-Y-Numpad(s).  For example:  b-2-3-5",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "wind",
					Description: "Wind direction.  In the form of compass directions.  For example:  SWW",
					Required:    false,
				},
			},
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){

		"arty": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			log.Printf("Got a command")
			var response string
			var wind string
			var from, to coord
			var azimuth float64
			var distance int

			var windString string
			var windDistString string
			var windAzString string

			fromString := i.ApplicationCommandData().Options[0].StringValue()
			from, err := NewCoord(fromString)
			if err != nil {
				sendResp("Error when calculating, make sure your inputs are correct.", s, i)
				return
			}

			toString := i.ApplicationCommandData().Options[1].StringValue()
			to, err = NewCoord(toString)
			if err != nil {
				sendResp("Error when calculating, make sure your inputs are correct.", s, i)
				return
			}

			azimuth, distance, err = calcArty(from, to)
			if err != nil {
				sendResp("Error when calculating, make sure your inputs are correct.", s, i)
				return
			}

			if len(i.ApplicationCommandData().Options) >= 3 {
				//We have wind
				wind = strings.ToUpper(i.ApplicationCommandData().Options[2].StringValue())
				windDir := getWindDir(wind)
				for _, offset := range WIND_OFFSETS {
					//Negative offsets since we want to go the opposite way.
					offsetCoord := offsetCoord(to, windDir, -offset)
					a, d, err := calcArty(from, offsetCoord)
					if err != nil {
						sendResp("Error when calculating, make sure your inputs are correct.", s, i)
						return
					}
					windString += fmt.Sprintf(" %4dM  |", offset)
					windDistString += fmt.Sprintf(" %5d  |", d)
					windAzString += fmt.Sprintf(" %5.1f  |", a)
				}
			} else {
				wind = "none"
			}

			response = fmt.Sprintf("```\nFrom: %s\nTo:   %s\nWind: %s\n\n"+
				"-----------------------------------------------------------------------------------\n"+
				"Wind:      none  | %s\n"+
				"-----------------------------------------------------------------------------------\n"+
				"Distance: %5d  | %s\n"+
				"Asimuth:  %5.1f  | %s\n"+
				"-----------------------------------------------------------------------------------"+
				"```", fromString, toString, wind, windString, distance, windDistString, azimuth, windAzString)

			sendResp(response, s, i)

		},
	}
)

func sendResp(response string, s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		// Ignore type for now, we'll discuss them in "responses" part
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
}

func numpadToPos(numpad int) (int, int) {
	switch numpad {
	case 7:
		return -1, -1
	case 8:
		return -0, -1
	case 9:
		return 1, -1
	case 4:
		return -1, 0
	case 5:
		return 0, 0
	case 6:
		return 1, 0
	case 1:
		return -1, 1
	case 2:
		return 0, 1
	case 3:
		return 1, 1
	}
	return 0, 0
}

type coord struct {
	x int
	y int
}

func NewCoord(rawString string) (c coord, err error) {

	log.Printf("Creating coord: %s", rawString)
	//Move our starting coord to the center of a tile.
	c.x = TILE_SIZE / 2
	c.y = TILE_SIZE / 2
	splitCoord := strings.Split(rawString, "-")

	c.x += charToInt(rune(splitCoord[0][0])) * TILE_SIZE
	yInt, err := strconv.Atoi(splitCoord[1])
	if err != nil {
		return
	}
	c.y += yInt * TILE_SIZE

	var numpadNum int

	offset := SMALL_TILE_SIZE
	for i := 2; i < len(splitCoord); i++ {
		numpadNum, err = strconv.Atoi(splitCoord[i])
		if err != nil {
			return
		}
		log.Printf("numpad num: %d", numpadNum)
		x, y := numpadToPos(numpadNum)
		c.x += x * offset
		c.y += y * offset
		offset /= 3
	}
	//Reverse since foxhole counts downwards.
	c.y = -c.y

	log.Printf("x: %d, y: %d", c.x, c.y)
	return

}

func (c coord) Subtract(s coord) (new coord) {
	new.x = c.x - s.x
	new.y = c.y - s.y
	return
}

func getWindDir(input string) (dir float64) {

	//Remember, in normal coords to the right is 0.  So we have to take that into account.
	//We transform it at the end.
	for _, char := range input {
		switch char {
		case 'S':
			dir += 90 * (math.Pi / 180)
		case 'W':
			dir += 180 * (math.Pi / 180)
		case 'N':
			dir += 270 * (math.Pi / 180)
		case 'E':
			dir += 0
		}
	}
	dir /= float64(len(input))
	return
}

func offsetCoord(c coord, dir float64, length int) (offsetCoord coord) {
	offsetCoord.x = c.x
	offsetCoord.y = c.y
	offsetCoord.x += int(math.Cos(dir) * float64(length))
	offsetCoord.y -= int(math.Sin(dir) * float64(length))
	return
}

func calcArty(from coord, to coord) (azimuth float64, distance int, err error) {

	log.Printf("from: %v, to: %v", from, to)

	//We normalize to from being zero
	n := to.Subtract(from)

	log.Printf("Normalized to zero: %v, %v", n.x, n.y)
	//A squared plus B squared equals C squared!
	distance = int(math.Round(math.Abs(math.Sqrt(float64(n.x*n.x) + float64(n.y*n.y)))))
	azimuth = math.Atan2(float64(n.y), float64(n.x)) * (180 / math.Pi)

	//Azimuth is off by 90 degrees
	azimuth -= 90

	//This is due to weirdness.  Makes the angles correct.
	azimuth = -azimuth
	for azimuth < 0 {
		azimuth += 360
	}
	//otherwise the printing gets weird....
	azimuth = math.Abs(azimuth)
	return
}

func charToInt(c rune) int {
	return int(unicode.ToUpper(c)-'A') + 1
}

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	var regCommands []*discordgo.ApplicationCommand
	for _, v := range commands {
		newCom, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		log.Printf("Added command: %v", v.Name)
		regCommands = append(regCommands, newCom)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
	}

	defer func() {
		log.Println("Gracefully shutdowning")
		/*if *RemoveCommands {
			for _, v := range regCommands {
				err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
				log.Printf("Deleted command: %v", v.Name)
				if err != nil {
					log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
				}
			}
		}*/
		s.Close()
		log.Printf("Finished shutdown")
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	log.Println("Finshed startup")
	<-stop
	log.Printf("Shutting down")
}
