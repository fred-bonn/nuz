package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/fred-bonn/nuz/internal/engine"
	"github.com/fred-bonn/nuz/internal/pokeapi"
	"github.com/spf13/pflag"
)

const (
	defaultLearningIterations = 1000
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := pflag.NewFlagSet("nuz", pflag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [flags] <player_showdown> <opponent_showdown>\n\nExamples:\n  %s -p -i 250 player.txt opponent.txt\n  %s -f policies/player__vs__opponent.json -i 1\n\n", os.Args[0], os.Args[0], os.Args[0])
		fs.PrintDefaults()
	}

	verbose := fs.BoolP("verbose", "v", false, "verbose logging")
	weather := fs.IntP("weather", "w", 0, "weather\n 0: None (default)\n 1: Rain\n 2: Sun\n 3: Sandstorm\n 4: Hail")
	inputAi := fs.IntP("input-ai", "a", 0, "input AI\n 0: Run & Bun (default)\n 1: Learning\n 2: Guided\n 3: Random")
	policyFile := fs.StringP("policy-file", "f", "", "path to a saved policy JSON file to load and use for the player trainer; the player and opponent parties embedded in the policy are used")
	iterations := fs.IntP("iterations", "i", 1, "number of times to run the same battle scenario for statistics or training")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return 0
		}
		log.Printf("error: invalid flags: %s", err)
		return 1
	}
	if *weather < 0 || *weather > 4 {
		log.Printf("error: weather (-w) must be between 0 and 4")
		return 1
	}
	if *iterations <= 0 {
		log.Printf("error: iterations must be greater than 0")
		return 1
	}
	if *inputAi < 0 || *inputAi > 3 {
		log.Printf("error: input AI (-a) must be between 0 and 3")
		return 1
	}
	if *inputAi == 1 && *iterations == 1 {
		*iterations = defaultLearningIterations
	}
	if *verbose {
		engine.Verbose = true
	}

	learning := *inputAi == 1

	parsedArgs := fs.Args()

	var policy *engine.SavedPolicy
	usingPolicyFile := *policyFile != ""
	if !usingPolicyFile {
		if len(parsedArgs) != 2 {
			log.Printf("error: missing arguments: usage: <executable> <player_showdown> <opponent_showdown> <flags>")
			return 1
		}
	} else {
		var err error
		policy, err = engine.LoadPolicyFromDisk(*policyFile)
		if err != nil {
			log.Printf("error: failed loading policy '%s': %s", *policyFile, err)
			return 1
		}
		if policy.PlayerParty == "" || policy.OpponentParty == "" {
			log.Printf("error: policy '%s' does not contain embedded party files", *policyFile)
			return 1
		}
	}

	cfg := &config{
		client: pokeapi.NewClient(),
	}

	var err error
	var playerParty, opponentParty []*engine.Pokemon
	if !usingPolicyFile {
		playerFileData, err := os.ReadFile(parsedArgs[0])
		if err != nil {
			log.Printf("error: failed reading player party file '%s': %s", parsedArgs[0], err)
			return 1
		}
		playerParty, err = cfg.validateInput(string(playerFileData))
		if err != nil {
			log.Printf("error: failed validating player party '%s': %s", parsedArgs[0], err)
			return 1
		}
		opponentFileData, err := os.ReadFile(parsedArgs[1])
		if err != nil {
			log.Printf("error: failed reading opponent party file '%s': %s", parsedArgs[0], err)
			return 1
		}
		opponentParty, err = cfg.validateInput(string(opponentFileData))
		if err != nil {
			log.Printf("error: failed validating player party '%s': %s", parsedArgs[0], err)
			return 1
		}
	} else {
		playerParty, err = cfg.validateInput(policy.PlayerParty)
		if err != nil {
			log.Printf("error: failed validating player party embedded in policy '%s': %s", *policyFile, err)
			return 1
		}
		opponentParty, err = cfg.validateInput(policy.OpponentParty)
		if err != nil {
			log.Printf("error: failed validating opponent party embedded in policy '%s': %s", *policyFile, err)
			return 1
		}
	}

	var playerAI engine.AI

	if usingPolicyFile {
		playerAI = engine.NewStaticPolicyAIFromPolicy(policy)
	} else {
		switch *inputAi {
		case 0:
			playerAI = engine.RnbAi{}
		case 1:
			playerAI = engine.NewLearningAI()
		case 2:
			engine.Verbose = true
			playerAI = engine.NewGuidedAI(os.Stdin, os.Stdout)
		case 3:
			playerAI = engine.RandomAi{}
		}
	}

	var bs engine.BattleState = engine.InitSingleBattleState(
		engine.Trainer{
			AI:           playerAI,
			Player:       true,
			FieldEffects: make(map[engine.FieldEffect]int),
		},
		engine.Trainer{
			AI:           engine.RnbAi{},
			FieldEffects: make(map[engine.FieldEffect]int),
		},
		playerParty,
		opponentParty,
		engine.WeatherState(*weather),
	)

	for i := 0; i < *iterations; i++ {
		if err := bs.Reset(); err != nil {
			log.Fatal(err)
		}
		if err := bs.Execute(); err != nil {
			log.Fatal(err)
		}
		bs.RecordStatistics()
		if learning, ok := bs.(*engine.SingleBattleState); ok {
			if playerAI, ok := learning.Player.AI.(*engine.LearningAI); ok {
				playerAI.RecordBattleOutcome(learning.GetStatistics())
				if playerAI.PolicySaturated() {
					log.Printf("training policy saturated after %d battle(s)", i+1)
					break
				}
				if learning.GetStatistics().AllPartySurvived() {
					log.Printf("training target reached: all party members survived after %d battle(s)", i+1)
					break
				}
			}
		}
	}

	if learning {
		if err := engine.SavePolicyToDisk((playerAI.(*engine.LearningAI)), parsedArgs[0], parsedArgs[1]); err != nil {
			log.Printf("error: failed saving policy: %s", err)
		} else {
			log.Printf("policy saved to %s", engine.PolicyPathForInputs(parsedArgs[0], parsedArgs[1]))
		}
	}

	if *iterations > 1 {
		bs.PrintStatistics()
	}

	return 0
}
