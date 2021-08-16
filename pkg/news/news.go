package news

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type Service struct {
	pool *pgxpool.Pool
}

type News struct {
	Id      int64
	Title   string
	Text    string
	Created int64
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Save(ctx context.Context, title, text string) error {
 	_, err := s.pool.Exec(ctx, `INSERT INTO news(title, text) VALUES($1, $2)`, title, text)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *Service) GetLatest(ctx context.Context) ([]*News, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, title, text, EXTRACT(EPOCH FROM created)::INT FROM news ORDER BY created DESC LIMIT 5`,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	latestNews := make([]*News, 0)
	for rows.Next() {
		var news News
		err = rows.Scan(&news.Id, &news.Title, &news.Text, &news.Created)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		latestNews = append(latestNews, &news)
	}

	return latestNews, nil
}
