---
name: 미국주식 데이터 소스 선택
description: 미국주식 히스토리컬 데이터 수집의 primary 소스로 Tiingo API를 선택한다.
status: 수락됨
date: 2026-02-14
supersedes: null
---

# ADR-0001: 미국주식 데이터 소스 선택

## 맥락

52주 일봉 OHLCV + adj_close 데이터를 수집해야 한다. 워치리스트 ~20종목, 수동 트리거 방식의 개인용 CLI 도구. 한국주식은 KIS OpenAPI(공식)로 확정. 미국주식 소스를 선정해야 한다.

## 결정 요인

- Go `net/http`로 직접 구현 가능한 단순한 인증 방식
- 무료 티어에서 ~20종목 수동 실행이 가능한 충분한 호출 한도
- adj_close 포함 일봉 데이터 제공
- API 안정성 (공식 문서, SLA 존재 여부)

## 검토한 선택지

### 선택지 1: Yahoo Finance

- 장점: 무료, 한국주식(.KS/.KQ)도 지원, 가장 널리 알려진 소스
- 단점: 비공식 API(SLA 없음), Cookie/Crumb + TLS 핑거프린트 스푸핑 필요, 2023년 이후 연 2-3회 인증 방식 변경으로 라이브러리 파손 반복, Go 라이브러리 미성숙(go-yfinance Stars 8)

### 선택지 2: Tiingo

- 장점: 공식 REST API, API key 인증(헤더 1줄), adj_close 기본 제공, 30년+ 히스토리, 무료 티어(50 req/hr, 500 symbols/mo)로 충분
- 단점: 1인 기업(bus factor=1), rate limit 초과 시 HTTP 200 + 비-JSON 응답 반환(429 아님), 한국주식 미지원

### 선택지 3: Alpha Vantage

- 장점: 공식 API, 글로벌 커버리지
- 단점: 무료 티어 25 req/day — 20종목 백필 시 수일 소요, adj_close 별도 엔드포인트

### 선택지 4: Polygon.io

- 장점: 공식 API, 실시간 데이터, 높은 안정성
- 단점: 무료 티어 5 req/min + 2년 히스토리 제한, 52주 데이터 수집에 부적합

## 결정

**Tiingo API를 미국주식 primary 소스로 선택한다.** 무료 티어 호출 한도가 사용량에 충분하고, API key 인증만으로 구현이 단순하다.

소스 교체에 대비해 `StockDataFetcher` 인터페이스를 분리한다. 환율(USD/KRW)은 Tiingo의 KRW 미지원 가능성이 높아 Frankfurter API를 별도 사용한다.

## 결과

- 긍정적: `net/http`만으로 구현 가능, 인증 파손 리스크 없음, adj_close 기본 제공
- 수용한 트레이드오프: 1인 기업 안정성 리스크(인터페이스 분리로 완화), rate limit 응답 바디 파싱 필요, 환율 별도 수집 필요

## 관련 문서

- [tiingo-python#45](https://github.com/hydrosquall/tiingo-python/issues/45) — rate limit 비-JSON 응답 이슈
- [Frankfurter API](https://frankfurter.dev/) — 환율 수집 소스
