// Package visibility applies the ETL's hidden-token policy to API reads.
// skip_parse is deliberately independent: hiding must not delete indexed data.
package visibility

import (
	"fmt"
	"strings"
)

// All aliases passed to these helpers are source-code constants, never request input.
func Token(alias string) string {
	return fmt.Sprintf(`NOT EXISTS (SELECT 1 FROM token_exception hidden_token
 WHERE hidden_token.chain_id = %[1]s.chain_id AND hidden_token.hidden
 AND hidden_token.contract = %[1]s.address)`, alias)
}

func Assets(alias string) string {
	return fmt.Sprintf(`NOT EXISTS (SELECT 1 FROM token_exception hidden_token
 WHERE hidden_token.chain_id = %[1]s.chain_id AND hidden_token.hidden
 AND (hidden_token.contract = %[1]s.asset0 OR hidden_token.contract = %[1]s.asset1))`, alias)
}

func Route(alias string) string {
	return fmt.Sprintf(`%[1]s.deleted_at IS NULL AND NOT EXISTS (
 SELECT 1 FROM token_exception hidden_token
 WHERE hidden_token.chain_id = %[1]s.chain_id AND hidden_token.hidden
 AND (hidden_token.contract = %[1]s.asset0 OR hidden_token.contract = %[1]s.asset1
 OR hidden_token.contract = ANY(%[1]s.route)))`, alias)
}

// Price excludes historical quotes whose route is retired or reaches a hidden token.
func Price(alias string) string {
	return fmt.Sprintf("EXISTS (SELECT 1 FROM route visible_route WHERE visible_route.id = %s.route_id AND visible_route.chain_id = %s.chain_id AND %s)", alias, alias, Route("visible_route"))
}

// Query gives reporting SQL a consistent view of visible data, including nested
// aggregates and inactive-pool fallbacks. Non-recursive CTEs read the physical
// table in their own definition, then shadow it for subsequent relations.
// NOT MATERIALIZED lets PostgreSQL push each report's chain/time filters down
// instead of materializing complete history tables referenced more than once.
// This accepts application SELECT statements, optionally starting with WITH.
func Query(query string) string {
	query = strings.TrimSpace(query)
	prefix := `WITH
 tokens AS NOT MATERIALIZED (SELECT * FROM tokens WHERE ` + Token("tokens") + `),
 pair AS NOT MATERIALIZED (SELECT * FROM pair WHERE ` + Assets("pair") + `),
 route AS NOT MATERIALIZED (SELECT * FROM route WHERE ` + Route("route") + `),
 parsed_tx AS NOT MATERIALIZED (SELECT * FROM parsed_tx WHERE ` + Assets("parsed_tx") + `),
 pair_stats_30m AS NOT MATERIALIZED (
 SELECT * FROM pair_stats_30m s WHERE EXISTS
 (SELECT 1 FROM pair p WHERE p.chain_id = s.chain_id AND p.id = s.pair_id)),
 pair_stats_recent AS NOT MATERIALIZED (
 SELECT * FROM pair_stats_recent s WHERE EXISTS
 (SELECT 1 FROM pair p WHERE p.chain_id = s.chain_id AND p.id = s.pair_id)),
 price AS NOT MATERIALIZED (
 SELECT * FROM price p WHERE EXISTS
 (SELECT 1 FROM route r WHERE r.chain_id = p.chain_id AND r.id = p.route_id))
 `
	if len(query) >= 5 && strings.EqualFold(query[:4], "with") && strings.ContainsAny(query[4:5], " \n\t\r") {
		return prefix + ", " + strings.TrimSpace(query[4:])
	}
	return prefix + query
}
