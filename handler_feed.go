package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gauravpatil2468/rssagg/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (apiCfg *apiConfig) handlerCreateFeed(w http.ResponseWriter, r *http.Request, user db.User) {
	type parameters struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Errror parsing JSON: %v", err))
		return
	}
	feed, err := apiCfg.DB.CreateFeed(r.Context(), db.CreateFeedParams{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		CreatedAt: pgtype.Timestamp{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamp{
			Time:  time.Now(),
			Valid: true,
		},
		Name:   params.Name,
		Url:    params.URL,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldnt create feed: %s", err))
	}
	respondWithJson(w, 201, databaseFeedToFeed(feed))
}

func (apiCfg *apiConfig) handlerGetAllFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := apiCfg.DB.GetFeeds(r.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get feeds: %v", err))
		return
	}
	respondWithJson(w, 200, databaseFeedsToFeeds(feeds))
}
