package pokeapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() Client {
	return Client{
		httpClient: &http.Client{},
		baseURL:    "https://pokeapi.co/api/v2",
	}
}

func (c *Client) fetchJSON(ctx context.Context, url string, v any) error {
	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}
	if c.baseURL == "" {
		c.baseURL = "https://pokeapi.co/api/v2"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error building API request: %w", err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching API data: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("api request failed with status %s", res.Status)
	}

	if err := json.NewDecoder(res.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

func (c *Client) FetchPokemon(ctx context.Context, name string) (PokemonJSON, error) {
	url := fmt.Sprintf("%s/pokemon/%s", c.baseURL, name)
	var pokemonJSON PokemonJSON
	if err := c.fetchJSON(ctx, url, &pokemonJSON); err != nil {
		return PokemonJSON{}, fmt.Errorf("error fetching Pokemon data from API: %w", err)
	}

	pokemonJSON.Name = cleanName(pokemonJSON.Name)

	return pokemonJSON, nil
}

func (c *Client) FetchMove(ctx context.Context, name string) (MoveJSON, error) {
	url := fmt.Sprintf("%s/move/%s", c.baseURL, name)
	var move MoveJSON
	if err := c.fetchJSON(ctx, url, &move); err != nil {
		return MoveJSON{}, fmt.Errorf("error fetching Move data from API: %w", err)
	}

	move.Name = cleanName(move.Name)

	return move, nil
}

func cleanName(name string) string {
	name = strings.ToLower(name)
	if !hasHyphen(name) && !isRegionalPokemon(name) {
		name = strings.ReplaceAll(name, "-", " ")
	}

	return name
}

func hasHyphen(name string) bool {
	var withHyphen = map[string]struct{}{
		"ho-oh":     {},
		"porygon-z": {},
		"jangmo-o":  {},
		"hakamo-o":  {},
		"kommo-o":   {},
		"ting-lu":   {},
		"chien-pao": {},
		"wo-chien":  {},
		"chi-yu":    {},
	}

	if _, ok := withHyphen[name]; ok {
		return true
	}

	return false
}

func isRegionalPokemon(name string) bool {
	regions := []string{
		"-alola",
		"-galar",
		"-hisui",
		"-paldea",
	}

	name = strings.ToLower(name)

	for _, region := range regions {
		if strings.HasSuffix(name, region) {
			return true
		}
	}

	return false
}
