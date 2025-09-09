package short

import "go.mongodb.org/mongo-driver/bson/primitive"

type shortLinkDoc struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Key         string             `bson:"key"`
	OriginalURL string             `bson:"originalURL"`
	ShortURL    string             `bson:"shortURL"`
	Hits        uint               `bson:"hits"`
	CreatedAt   primitive.DateTime `bson:"createdAt"`
}

type statsDoc struct {
	Hits      uint               `bson:"hits"`
	CreatedAt primitive.DateTime `bson:"createdAt"`
}
