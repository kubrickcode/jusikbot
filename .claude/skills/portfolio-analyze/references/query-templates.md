# PostgreSQL Query Templates

Conditionally loaded when `data/summary.md` is insufficient. Connect via `$DATABASE_URL`.

## Detailed Price History

```sql
SELECT date, open, high, low, close, adj_close, volume
FROM price_history
WHERE symbol = $1 AND date >= CURRENT_DATE - INTERVAL '52 weeks'
ORDER BY date;
```

## Moving Average Calculation

```sql
SELECT date, adj_close,
  AVG(adj_close) OVER (ORDER BY date ROWS 49 PRECEDING) AS ma_50,
  AVG(adj_close) OVER (ORDER BY date ROWS 199 PRECEDING) AS ma_200
FROM price_history
WHERE symbol = $1 AND date >= CURRENT_DATE - INTERVAL '60 weeks'
ORDER BY date DESC
LIMIT 20;
```

## Volume Analysis

```sql
SELECT date, volume,
  AVG(volume) OVER (ORDER BY date ROWS 19 PRECEDING) AS avg_vol_20d
FROM price_history
WHERE symbol = $1 AND date >= CURRENT_DATE - INTERVAL '8 weeks'
ORDER BY date DESC
LIMIT 5;
```

## Relative Strength Comparison

```sql
WITH returns AS (
  SELECT symbol,
    (LAST_VALUE(adj_close) OVER w - FIRST_VALUE(adj_close) OVER w)
      / FIRST_VALUE(adj_close) OVER w AS return_pct
  FROM price_history
  WHERE symbol IN ($1, $2) AND date >= CURRENT_DATE - ($3)::interval
  WINDOW w AS (PARTITION BY symbol ORDER BY date ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING)
)
SELECT DISTINCT symbol, return_pct FROM returns;
```

## Drawdown from 52W High

```sql
SELECT symbol,
  MAX(adj_close) AS high_52w,
  (SELECT adj_close FROM price_history p2 WHERE p2.symbol = p1.symbol ORDER BY date DESC LIMIT 1) AS latest,
  1 - (SELECT adj_close FROM price_history p2 WHERE p2.symbol = p1.symbol ORDER BY date DESC LIMIT 1) / MAX(adj_close) AS drawdown_pct
FROM price_history p1
WHERE date >= CURRENT_DATE - INTERVAL '52 weeks'
GROUP BY symbol;
```

## RSI Approximation (14-day SMA)

```sql
WITH daily AS (
  SELECT date, adj_close,
    adj_close - LAG(adj_close) OVER (ORDER BY date) AS change
  FROM price_history
  WHERE symbol = $1 AND date >= CURRENT_DATE - INTERVAL '6 weeks'
),
avg_gains AS (
  SELECT
    AVG(CASE WHEN change > 0 THEN change ELSE 0 END) AS avg_gain,
    AVG(CASE WHEN change < 0 THEN ABS(change) ELSE 0 END) AS avg_loss
  FROM daily
  WHERE date >= CURRENT_DATE - INTERVAL '14 days'
)
SELECT 100 - (100 / (1 + avg_gain / NULLIF(avg_loss, 0))) AS rsi_14 FROM avg_gains;
```

## KRW Conversion

```sql
SELECT p.symbol, p.date, p.adj_close,
  p.adj_close * f.rate AS value_krw
FROM price_history p
JOIN fx_rate f ON f.pair = 'USDKRW' AND f.date = (
  SELECT MAX(date) FROM fx_rate WHERE pair = 'USDKRW' AND date <= p.date
)
WHERE p.symbol = $1
ORDER BY p.date DESC
LIMIT 5;
```

## Multi-Symbol Technical Summary

```sql
SELECT symbol,
  (SELECT adj_close FROM price_history p2 WHERE p2.symbol = p1.symbol ORDER BY date DESC LIMIT 1) AS latest,
  MAX(adj_close) AS high_52w,
  MIN(adj_close) AS low_52w,
  AVG(adj_close) FILTER (WHERE date >= CURRENT_DATE - INTERVAL '50 days') AS ma_50
FROM price_history p1
WHERE date >= CURRENT_DATE - INTERVAL '52 weeks'
GROUP BY symbol
ORDER BY symbol;
```
