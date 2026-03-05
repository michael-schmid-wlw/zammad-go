package zammad

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Ticket is a zammad ticket.
type Ticket[T any] struct {
	Title                 string        `json:"title"`
	Group                 string        `json:"group"`
	OwnerID               int           `json:"owner_id,omitempty"`
	ID                    int           `json:"id,omitempty"`
	Article               TicketArticle `json:"article,omitempty"`
	GroupID               int           `json:"group_id,omitempty"`
	PriorityID            int           `json:"priority_id,omitempty"`
	StateID               int           `json:"state_id,omitempty"`
	State                 string        `json:"state,omitempty"`
	OrganizationID        int           `json:"organization_id"`
	Number                string        `json:"number,omitempty"`
	Customer              string        `json:"customer,omitempty"`
	CustomerID            int           `json:"customer_id,omitempty"`
	LastContactAt         time.Time     `json:"last_contact_at,omitempty"`
	LastContactAgentAt    time.Time     `json:"last_contact_agent_at,omitempty"`
	LastContactCustomerAt time.Time     `json:"last_contact_customer_at,omitempty"`
	CreateArticleTypeID   int           `json:"create_article_type_id,omitempty"`
	CreateArticleSenderID int           `json:"create_article_sender_id,omitempty"`
	ArticleCount          int           `json:"article_count,omitempty"`
	UpdatedByID           int           `json:"updated_by_id,omitempty"`
	CreatedByID           int           `json:"created_by_id,omitempty"`
	CreatedAt             time.Time     `json:"created_at,omitempty"`
	UpdatedAt             time.Time     `json:"updated_at,omitempty"`
	CustomFields          T             `json:"-"`
}

func (t *Ticket[T]) UnmarshalJSON(data []byte) error {
	type TicketAlias Ticket[T]
	if err := json.Unmarshal(data, (*TicketAlias)(t)); err != nil {
		return err
	}
	return json.Unmarshal(data, &t.CustomFields)
}

func (t Ticket[T]) MarshalJSON() ([]byte, error) {
	type TicketAlias Ticket[T]

	base, err := json.Marshal(TicketAlias(t))
	if err != nil {
		return nil, err
	}

	custom, err := json.Marshal(t.CustomFields)
	if err != nil {
		return nil, err
	}

	// T=struct{} oder keine Custom Fields gesetzt → base direkt zurückgeben
	if len(custom) <= 2 {
		return base, nil
	}

	// {"title":"..."} + {"cf_foo":"bar"} → {"title":"...","cf_foo":"bar"}
	merged := make([]byte, 0, len(base)+len(custom))
	merged = append(merged, base[:len(base)-1]...) // trailing } entfernen
	merged = append(merged, ',')
	merged = append(merged, custom[1:]...) // leading { entfernen
	return merged, nil
}

func (c *client[T]) TicketListResult(opts ...Option) *Result[Ticket[T]] {
	return &Result[Ticket[T]]{
		res:     nil,
		resFunc: c.TicketListWithOptions,
		opts:    NewRequestOptions(opts...),
	}
}

func (c *client[T]) TicketList() ([]Ticket[T], error) {
	return c.TicketListResult().FetchAll()
}

func (c *client[T]) TicketListWithOptions(ro RequestOptions) ([]Ticket[T], error) {
	var tickets []Ticket[T]

	req, err := c.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", c.Url, "/api/v1/tickets"), nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = ro.URLParams()

	if err = c.sendWithAuth(req, &tickets); err != nil {
		return nil, err
	}

	return tickets, nil
}

// TicketSearch searches for tickets. See https://docs.zammad.org/en/latest/api/ticket/index.html#search.
func (c *client[T]) TicketSearch(query string, limit int) ([]Ticket[T], error) {
	type Assets struct {
		AssetTicket map[int]Ticket[T] `json:"ticket"`
	}

	type TickSearch struct {
		Tickets []int `json:"tickets"`
		Count   int   `json:"tickets_count"`
		Assets  `json:"assets"`
	}

	var ticksearch TickSearch
	req, err := c.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", c.Url, fmt.Sprintf("/api/v1/tickets/search?query=%s&limit=%d", url.QueryEscape(query), limit)), nil)
	if err != nil {
		return nil, err
	}

	if err = c.sendWithAuth(req, &ticksearch); err != nil {
		return nil, err
	}

	tickets := make([]Ticket[T], ticksearch.Count)
	i := 0
	for _, t1 := range ticksearch.Assets.AssetTicket {
		tickets[i] = t1
		i++
	}
	return tickets, nil
}

func (c *client[T]) TicketShow(ticketID int) (Ticket[T], error) {
	var ticket Ticket[T]

	req, err := c.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", c.Url, fmt.Sprintf("/api/v1/tickets/%d", ticketID)), nil)
	if err != nil {
		return ticket, err
	}

	if err = c.sendWithAuth(req, &ticket); err != nil {
		return ticket, err
	}

	return ticket, nil
}

// TicketCreate is used to create a ticket. For this you need to assemble a bare-bones Ticket:
//
//	ticket := Ticket{
//		Title:      "your subject",
//		Group:      "your group",
//		CustomerID: 10, // your customer ID
//		Article: TicketArticle{
//			Subject: "subject of comment",
//			Body: "body of comment",
//		},
//	}
func (c *client[T]) TicketCreate(t Ticket[T]) (Ticket[T], error) {
	var ticket Ticket[T]

	req, err := c.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", c.Url, "/api/v1/tickets"), t)
	if err != nil {
		return ticket, err
	}

	if err = c.sendWithAuth(req, &ticket); err != nil {
		return ticket, err
	}

	return ticket, nil
}

func (c *client[T]) TicketUpdate(ticketID int, t Ticket[T]) (Ticket[T], error) {
	var ticket Ticket[T]

	req, err := c.NewRequest(http.MethodPut, fmt.Sprintf("%s%s", c.Url, fmt.Sprintf("/api/v1/tickets/%d", ticketID)), t)
	if err != nil {
		return ticket, err
	}

	if err = c.sendWithAuth(req, &ticket); err != nil {
		return ticket, err
	}

	return ticket, nil
}

func (c *client[T]) TicketDelete(ticketID int) error {
	req, err := c.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s", c.Url, fmt.Sprintf("/api/v1/tickets/%d", ticketID)), nil)
	if err != nil {
		return err
	}

	if err = c.sendWithAuth(req, nil); err != nil {
		return err
	}

	return nil
}
