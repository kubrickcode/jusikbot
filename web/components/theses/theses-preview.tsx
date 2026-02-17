"use client"

import type { Components } from "react-markdown"
import Markdown from "react-markdown"
import remarkGfm from "remark-gfm"

type ThesesPreviewProps = {
  content: string
}

const markdownComponents: Components = {
  h1: ({ children }) => (
    <h1 className="text-2xl font-bold mt-8 mb-4 pb-2 border-b border-border first:mt-0">
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2 className="text-xl font-semibold mt-8 mb-3 text-foreground">{children}</h2>
  ),
  h3: ({ children }) => (
    <h3 className="text-base font-semibold mt-4 mb-2 text-foreground">{children}</h3>
  ),
  p: ({ children }) => (
    <p className="text-sm leading-relaxed mb-3 text-foreground">{children}</p>
  ),
  strong: ({ children }) => (
    <strong className="font-semibold text-foreground">{children}</strong>
  ),
  blockquote: ({ children }) => (
    <blockquote className="border-l-2 border-primary pl-4 my-4 text-muted-foreground text-sm italic">
      {children}
    </blockquote>
  ),
  code: ({ children, className }) => {
    const isBlock = className?.includes("language-")
    if (isBlock) {
      return (
        <code className="block bg-muted rounded-md p-4 text-xs font-mono leading-relaxed overflow-x-auto whitespace-pre">
          {children}
        </code>
      )
    }
    return <code className="bg-muted px-1.5 py-0.5 rounded text-xs font-mono">{children}</code>
  },
  pre: ({ children }) => <pre className="my-4">{children}</pre>,
  ul: ({ children }) => (
    <ul className="list-disc list-outside pl-5 space-y-1 mb-3 text-sm">{children}</ul>
  ),
  ol: ({ children }) => (
    <ol className="list-decimal list-outside pl-5 space-y-1 mb-3 text-sm">{children}</ol>
  ),
  li: ({ children }) => (
    <li className="text-sm text-foreground leading-relaxed">{children}</li>
  ),
  table: ({ children }) => (
    <div className="overflow-x-auto my-4">
      <table className="w-full text-sm border-collapse">{children}</table>
    </div>
  ),
  thead: ({ children }) => <thead>{children}</thead>,
  tbody: ({ children }) => <tbody>{children}</tbody>,
  tr: ({ children }) => <tr className="border-b border-border">{children}</tr>,
  th: ({ children }) => (
    <th className="border border-border px-3 py-2 text-left font-medium bg-muted text-xs">
      {children}
    </th>
  ),
  td: ({ children }) => (
    <td className="border border-border px-3 py-2 text-xs text-muted-foreground">
      {children}
    </td>
  ),
  a: ({ children, href }) => (
    <a
      href={href}
      className="text-primary underline underline-offset-2 hover:opacity-80"
      target="_blank"
      rel="noopener noreferrer"
    >
      {children}
    </a>
  ),
  hr: () => <hr className="my-6 border-border" />,
}

export function ThesesPreview({ content }: ThesesPreviewProps) {
  return (
    <div className="p-6 overflow-auto h-full">
      <Markdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
        {content}
      </Markdown>
    </div>
  )
}
