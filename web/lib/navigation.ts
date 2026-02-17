import {
  Sliders,
  List,
  Wallet,
  BookOpen,
  CandlestickChart,
  ArrowLeftRight,
  BarChart3,
  CheckCircle,
  Search,
  Clock,
  FileText,
  type LucideIcon,
} from "lucide-react"

type NavItem = {
  title: string
  href: string
  icon: LucideIcon
  description: string
}

type NavGroup = {
  label: string
  items: NavItem[]
}

export const navigationGroups: NavGroup[] = [
  {
    label: "설정",
    items: [
      {
        title: "예산 및 리스크",
        href: "/config/settings",
        icon: Sliders,
        description: "투자 예산, 리스크 한도, 포지션 사이징 설정",
      },
      {
        title: "추적 종목",
        href: "/config/watchlist",
        icon: List,
        description: "관심 종목 추가, 수정, 삭제",
      },
      {
        title: "보유 현황",
        href: "/config/holdings",
        icon: Wallet,
        description: "현재 포지션의 수량과 평균 단가",
      },
      {
        title: "투자 논제",
        href: "/config/theses",
        icon: BookOpen,
        description: "투자 논제와 유효/무효화 조건",
      },
    ],
  },
  {
    label: "시장 데이터",
    items: [
      {
        title: "가격 데이터",
        href: "/data/prices",
        icon: CandlestickChart,
        description: "종목별 OHLCV 가격 이력",
      },
      {
        title: "환율",
        href: "/data/fx",
        icon: ArrowLeftRight,
        description: "USD/KRW 환율 데이터",
      },
      {
        title: "기술 지표 요약",
        href: "/data/summary",
        icon: BarChart3,
        description: "14개 컬럼 기술 지표 요약 테이블",
      },
    ],
  },
  {
    label: "리서치 결과",
    items: [
      {
        title: "논제 팩트체크",
        href: "/research/thesis-check",
        icon: CheckCircle,
        description: "논제별 유효성 조건 충족 여부",
      },
      {
        title: "후보 종목",
        href: "/research/candidates",
        icon: Search,
        description: "리서치에서 발굴된 투자 후보",
      },
      {
        title: "리서치 히스토리",
        href: "/research/history",
        icon: Clock,
        description: "과거 팩트체크 결과 아카이브",
      },
    ],
  },
  {
    label: "리포트",
    items: [
      {
        title: "분석 리포트",
        href: "/reports",
        icon: FileText,
        description: "포트폴리오 분석 결과 보고서",
      },
    ],
  },
]
