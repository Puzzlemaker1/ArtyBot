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
	TILE_SIZE = 126
	SMALL_TILE_SIZE = 42
	WIND_OFFSET = 50.0
)

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

			from := i.ApplicationCommandData().Options[0].StringValue()
			to := i.ApplicationCommandData().Options[1].StringValue()
			var wind string
			if len(i.ApplicationCommandData().Options) >= 3 {
				//We have wind
				wind = i.ApplicationCommandData().Options[2].StringValue()
			}

			azimuth, distance, err := calcArty(from, to, wind)
			var response string
			if err != nil {
				response = "Error when calculating, make sure your inputs are correct."
			} else {
				response = fmt.Sprintf(`
	From: %s
	To: %s
	Distance: %d
	Asimuth:  %d
`, from, to, distance, azimuth)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				// Ignore type for now, we'll discuss them in "responses" part
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
		},

	}
)

func numpadToPos(numpad int)  (int, int) {
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

func calcArty(from string, to string, wind string) (azimuth int, distance int, err error) {
	fromCoords := strings.Split(from, "-")
	toCoords := strings.Split(to, "-")
	if len(fromCoords) < 3 || len(toCoords) < 3 {
		//Um.
		return
	}

	xfrom := charToInt(rune(fromCoords[0][0]))
	yfrom, err := strconv.Atoi(fromCoords[1])
	if err != nil {
		return
	}

	var numpadNum int

	var nxFrom, nyFrom int
	offset := SMALL_TILE_SIZE
	for i := 2; i < len(fromCoords); i++ {
		numpadNum, err = strconv.Atoi(fromCoords[i])
		if err != nil {
			return
		}
		x, y := numpadToPos(numpadNum)
		nxFrom += x * offset
		nyFrom += y * offset
		offset /= 3
	}

	xto := charToInt(rune(toCoords[0][0]))
	yto, err := strconv.Atoi(toCoords[1])
	if err != nil {
		return
	}

	var nxTo, nyTo int
	offset = SMALL_TILE_SIZE
	for i := 2; i < len(toCoords); i++ {
		numpadNum, err = strconv.Atoi(toCoords[i])
		if err != nil {
			return
		}
		x, y := numpadToPos(numpadNum)
		nxTo += x * offset
		nyTo += y * offset
		offset /= 3
	}





	log.Printf("xfrom: %v, yfrom: %v, xto: %v, yto: %v", xfrom, yfrom, xto, yto)
	log.Printf("nxfrom: %v, nyfrom: %v, nxto: %v, nyto: %v", nxFrom, nyFrom, nxTo, nyTo)

	//We normalize to meters, not co-ords
	xfrom *= TILE_SIZE
	xto *= TILE_SIZE
	yfrom *= TILE_SIZE
	yto *= TILE_SIZE

	//Add in our numpad options
	xfrom += nxFrom
	yfrom += nyFrom
	xto += nxTo
	yto += nyTo

	log.Printf("xfrom: %v, yfrom: %v, xto: %v, yto: %v", xfrom, yfrom, xto, yto)
	//NOw to calculate wind
	if wind != "" {
		var windDir float64

		//Remember, in normal coords to the right is 0.  So we have to take that into account.
		//We transform it at the end.
		for _, char := range wind {
			switch char {
			case 'S':
				windDir += 90 * (math.Pi/180)
			case 'W':
				windDir += 180 * (math.Pi/180)
			case 'N':
				windDir += 270 * (math.Pi/180)
			case 'E':
				windDir += 0
			}
		}
		windDir /= float64(len(wind))

		windOffsetX := math.Cos(windDir) * WIND_OFFSET
		windOffsetY := math.Sin(windDir) * WIND_OFFSET

		//Subtract since we do the opposite.
		xto -= int(windOffsetX)
		yto -= int(windOffsetY)
	}
	//Because foxhole is weird zero is upper left...
	yfrom = -yfrom
	yto = -yto

	//We normalize the "from" coord to zero
	x := float64(xto - xfrom)
	y := float64(yto - yfrom)



	log.Printf("Normalized to zero: %v, %v", x, y)
	//A squared plus B squared equals C squared!
	distance = int(math.Round(math.Abs(math.Sqrt((x * x) + (y * y)))))
	azimuth = int(math.Atan2(y, x) * (180 / math.Pi))


	//Azimuth is off by 90 degrees
	azimuth -= 90

	azimuth = azimuth % 360

	//This is due to weirdness.  Makes the angles correct.
	azimuth = -azimuth
	if azimuth < 0 {
		azimuth += 360
	}
	return
}

func charToInt(c rune) int {
	return int(unicode.ToUpper(c) - 'A') + 1
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
