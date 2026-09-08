package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/fred-bonn/nuz/internal/pokeapi"
	"github.com/spf13/pflag"
)

var verbose = pflag.BoolP("verbose", "v", false, "verbose logging")

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := pflag.NewFlagSet("nuzlocke-verifier", pflag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [flags] <player_showdown> <opponent_showdown>\n\nExamples:\n  %s -p -s -i 250 player.txt opponent.txt\n  %s -f policies/player__vs__opponent.json -i 1\n\n", os.Args[0], os.Args[0], os.Args[0])
		fs.PrintDefaults()
	}
	verbose = fs.BoolP("verbose", "v", false, "verbose logging")
	weather := fs.IntP("weather", "w", int(noneWeather), "weather\n 0: None (default)\n 1: Rain\n 2: Sun\n 3: Sandstorm\n 4: Hail")
	playerUsesLearningAI := fs.BoolP("player-learning-ai", "p", false, "use the learning AI for the player trainer while the opponent keeps the rnb AI")
	playerUsesGuidedAI := fs.BoolP("player-guided-ai", "g", false, "prompt for the player's action each turn")
	policyFile := fs.StringP("policy-file", "f", "", "path to a saved policy JSON file to load and use for the player trainer; the player and opponent parties embedded in the policy are used, so <player_showdown> <opponent_showdown> must not be given")
	savePolicy := fs.BoolP("save-policy", "s", false, "save the learned policy under policies/ using the input file names")
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

	parsedArgs := fs.Args()
	usingPolicyFile := *policyFile != ""
	if usingPolicyFile {
		if len(parsedArgs) != 0 {
			log.Printf("error: --policy-file uses the party files embedded in the policy; do not pass <player_showdown> <opponent_showdown>")
			return 1
		}
	} else if len(parsedArgs) != 2 {
		log.Printf("error: missing arguments: usage: <executable> <player_showdown> <opponent_showdown> <flags>")
		return 1
	}

	cfg := &config{
		client: pokeapi.NewClient(),
	}

	var policy *savedPolicy
	var playerParty, opponentParty []*pokemon
	if usingPolicyFile {
		var err error
		policy, err = loadPolicyFromDisk(*policyFile)
		if err != nil {
			log.Printf("error: failed loading policy '%s': %s", *policyFile, err)
			return 1
		}
		if policy.PlayerParty == "" || policy.OpponentParty == "" {
			log.Printf("error: policy '%s' does not contain embedded party files", *policyFile)
			return 1
		}
		playerParty, err = cfg.validateInputContent(policy.PlayerParty)
		if err != nil {
			log.Printf("error: failed validating player party embedded in policy '%s': %s", *policyFile, err)
			return 1
		}
		opponentParty, err = cfg.validateInputContent(policy.OpponentParty)
		if err != nil {
			log.Printf("error: failed validating opponent party embedded in policy '%s': %s", *policyFile, err)
			return 1
		}
	} else {
		var err error
		playerParty, err = cfg.validateInput(parsedArgs[0])
		if err != nil {
			log.Printf("error: failed validating input '%s': %s", parsedArgs[0], err)
			return 1
		}
		opponentParty, err = cfg.validateInput(parsedArgs[1])
		if err != nil {
			log.Printf("error: failed validating input '%s': %s", parsedArgs[1], err)
			return 1
		}
	}

	var playerLearning *learningAi
	playerAI := ai(rnbAi{})
	if *playerUsesGuidedAI {
		*verbose = true
		playerAI = newGuidedAi(os.Stdin, os.Stdout)
	} else if usingPolicyFile {
		if err := validatePolicyCompatibility(policy, playerParty, opponentParty); err != nil {
			log.Printf("error: policy incompatible with input parties: %s", err)
			return 1
		}
		playerAI = newStaticPolicyAiFromPolicy(policy)
		playerLearning = nil
		log.Printf("loaded policy from %s: %d states, %d scored actions, %d observed actions", *policyFile, len(policy.Policy), countScoreEntries(policy.Scores), countCountEntries(policy.Counts))
	} else if *playerUsesLearningAI {
		playerLearning = newLearningAi()
		playerAI = playerLearning
	}

	var bs battleState = initSingleBattleState(
		trainer{
			ai:           playerAI,
			player:       true,
			fieldEffects: make(map[fieldEffect]int),
		},
		trainer{
			ai:           rnbAi{},
			fieldEffects: make(map[fieldEffect]int),
		},
		playerParty,
		opponentParty,
		weatherState(*weather),
	)

	for i := 0; i < *iterations; i++ {
		if err := bs.reset(); err != nil {
			log.Fatal(err)
		}
		if err := bs.execute(); err != nil {
			log.Fatal(err)
		}
		bs.recordStatistics()
		if learning, ok := bs.(*singleBattleState); ok {
			if playerAI, ok := learning.player.ai.(*learningAi); ok {
				playerAI.recordBattleOutcome(learning.getStatistics())
				if playerAI.policySaturated() {
					log.Printf("training policy saturated after %d battle(s)", i+1)
					break
				}
				if learning.getStatistics().allPartySurvived() {
					log.Printf("training target reached: all party members survived after %d battle(s)", i+1)
					break
				}
			}
		}
	}

	if *savePolicy && playerLearning != nil {
		if err := savePolicyToDisk(playerLearning, parsedArgs[0], parsedArgs[1]); err != nil {
			log.Printf("error: failed saving policy: %s", err)
		} else {
			log.Printf("policy saved to %s", policyPathForInputs(parsedArgs[0], parsedArgs[1]))
		}
	}

	bs.printStatistics()
	return 0
}
