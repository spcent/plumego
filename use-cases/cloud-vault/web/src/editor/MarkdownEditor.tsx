import { forwardRef, useEffect, useImperativeHandle, useRef } from 'react'
import { EditorView, Decoration, keymap, lineNumbers, highlightActiveLine, ViewPlugin, type DecorationSet } from '@codemirror/view'
import { EditorState, RangeSetBuilder } from '@codemirror/state'
import { defaultKeymap, history, historyKeymap, redo, undo } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language'

export type MarkdownCommand =
  | 'heading'
  | 'bold'
  | 'italic'
  | 'code'
  | 'quote'
  | 'list'
  | 'orderedList'
  | 'task'
  | 'link'
  | 'image'
  | 'table'
  | 'undo'
  | 'redo'

export interface CursorPosition {
  line: number
  column: number
}

export interface MarkdownEditorHandle {
  applyCommand: (command: MarkdownCommand) => void
  focus: () => void
}

interface MarkdownEditorProps {
  value: string
  onChange: (value: string) => void
  onSave?: () => void
  highlightQuery?: string
  onCursorChange?: (position: CursorPosition) => void
}

const workspaceTheme = EditorView.theme({
  '&': {
    height: '100%',
    backgroundColor: 'hsl(var(--editor-bg))',
    color: 'hsl(var(--editor-text))',
  },
  '.cm-content': {
    padding: '18px 28px 40px 18px',
    fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
    fontSize: '13px',
    lineHeight: '1.72',
    caretColor: 'hsl(var(--editor-caret))',
  },
  '.cm-gutters': {
    backgroundColor: 'hsl(var(--editor-bg))',
    borderRight: '1px solid hsl(var(--workspace-border))',
    color: 'hsl(var(--editor-gutter))',
    fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
    fontSize: '13px',
    paddingLeft: '10px',
  },
  '.cm-activeLineGutter': {
    backgroundColor: 'hsl(var(--editor-active-line))',
    color: 'hsl(var(--workspace-text))',
  },
  '.cm-activeLine': {
    backgroundColor: 'hsl(var(--editor-active-line))',
  },
  '.cm-cursor': {
    borderLeftColor: 'hsl(var(--editor-caret))',
  },
  '.cm-selectionBackground': {
    backgroundColor: 'hsl(var(--editor-selection)) !important',
  },
  '.cm-search-highlight': {
    backgroundColor: 'hsl(var(--editor-highlight))',
    borderRadius: '2px',
  },
  '.cm-line': {
    paddingLeft: '8px',
  },
  '.tok-heading': {
    color: 'hsl(var(--workspace-accent))',
    fontWeight: '600',
  },
})

const searchHighlightMark = Decoration.mark({ class: 'cm-search-highlight' })

// buildHighlightPlugin creates a ViewPlugin that highlights all matches of `query`
// in the editor content. Re-runs on every query change via `provide`.
function buildHighlightPlugin(getQuery: () => string) {
  return ViewPlugin.define(
    () => ({
      decorations: Decoration.none as DecorationSet,
      update(update) {
        const q = getQuery().trim().toLowerCase()
        if (!q) {
          this.decorations = Decoration.none
          return
        }
        const builder = new RangeSetBuilder<Decoration>()
        const text = update.state.doc.toString().toLowerCase()
        let from = 0
        while (true) {
          const idx = text.indexOf(q, from)
          if (idx < 0) break
          builder.add(idx, idx + q.length, searchHighlightMark)
          from = idx + q.length
        }
        this.decorations = builder.finish()
      },
    }),
    { decorations: v => v.decorations },
  )
}

function selectionText(view: EditorView) {
  const selection = view.state.selection.main
  return view.state.doc.sliceString(selection.from, selection.to)
}

function replaceSelection(view: EditorView, insert: string, selectFrom?: number, selectTo?: number) {
  const selection = view.state.selection.main
  const base = selection.from
  view.dispatch({
    changes: { from: selection.from, to: selection.to, insert },
    selection: selectFrom != null && selectTo != null ? { anchor: base + selectFrom, head: base + selectTo } : { anchor: base + insert.length },
    scrollIntoView: true,
  })
  view.focus()
}

function prefixCurrentLine(view: EditorView, prefix: string) {
  const line = view.state.doc.lineAt(view.state.selection.main.head)
  const current = view.state.doc.sliceString(line.from, line.to)
  const existingIndent = current.match(/^\s*/)?.[0] ?? ''
  const insertAt = line.from + existingIndent.length
  view.dispatch({
    changes: { from: insertAt, insert: prefix },
    selection: { anchor: view.state.selection.main.head + prefix.length },
    scrollIntoView: true,
  })
  view.focus()
}

function applyMarkdownCommand(view: EditorView, command: MarkdownCommand) {
  const selected = selectionText(view)
  switch (command) {
    case 'heading':
      prefixCurrentLine(view, '# ')
      return
    case 'bold': {
      const text = selected || 'bold text'
      replaceSelection(view, `**${text}**`, 2, 2 + text.length)
      return
    }
    case 'italic': {
      const text = selected || 'italic text'
      replaceSelection(view, `_${text}_`, 1, 1 + text.length)
      return
    }
    case 'code': {
      const text = selected || 'code'
      replaceSelection(view, `\`${text}\``, 1, 1 + text.length)
      return
    }
    case 'quote':
      prefixCurrentLine(view, '> ')
      return
    case 'list':
      prefixCurrentLine(view, '- ')
      return
    case 'orderedList':
      prefixCurrentLine(view, '1. ')
      return
    case 'task':
      prefixCurrentLine(view, '- [ ] ')
      return
    case 'link': {
      const text = selected || 'link text'
      replaceSelection(view, `[${text}](https://)`, 1, 1 + text.length)
      return
    }
    case 'image':
      replaceSelection(view, '![alt text](image-url)', 2, 10)
      return
    case 'table':
      replaceSelection(view, '| Column | Value |\n|---|---|\n| Item | Detail |\n')
      return
    case 'undo':
      undo(view)
      return
    case 'redo':
      redo(view)
      return
  }
}

function cursorPosition(view: EditorView): CursorPosition {
  const head = view.state.selection.main.head
  const line = view.state.doc.lineAt(head)
  return { line: line.number, column: head - line.from + 1 }
}

export const MarkdownEditor = forwardRef<MarkdownEditorHandle, MarkdownEditorProps>(function MarkdownEditor({
  value,
  onChange,
  onSave,
  highlightQuery = '',
  onCursorChange,
}, ref) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const valueRef = useRef(value)
  const onChangeRef = useRef(onChange)
  const onSaveRef = useRef(onSave)
  const onCursorChangeRef = useRef(onCursorChange)
  const queryRef = useRef(highlightQuery)

  onChangeRef.current = onChange
  onSaveRef.current = onSave
  onCursorChangeRef.current = onCursorChange
  queryRef.current = highlightQuery

  useImperativeHandle(ref, () => ({
    applyCommand(command) {
      if (viewRef.current) applyMarkdownCommand(viewRef.current, command)
    },
    focus() {
      viewRef.current?.focus()
    },
  }), [])

  useEffect(() => {
    if (!containerRef.current) return

    const saveKeymap = keymap.of([
      {
        key: 'Mod-s',
        preventDefault: true,
        run: () => {
          onSaveRef.current?.()
          return true
        },
      },
    ])

    const highlightPlugin = buildHighlightPlugin(() => queryRef.current)

    const state = EditorState.create({
      doc: valueRef.current,
      extensions: [
        lineNumbers(),
        history(),
        highlightActiveLine(),
        syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
        markdown(),
        workspaceTheme,
        keymap.of([...defaultKeymap, ...historyKeymap]),
        saveKeymap,
        highlightPlugin,
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            const newValue = update.state.doc.toString()
            valueRef.current = newValue
            onChangeRef.current(newValue)
          }
          if (update.docChanged || update.selectionSet) {
            onCursorChangeRef.current?.(cursorPosition(update.view))
          }
        }),
        EditorView.lineWrapping,
      ],
    })

    const view = new EditorView({ state, parent: containerRef.current })
    viewRef.current = view
    onCursorChangeRef.current?.(cursorPosition(view))

    return () => {
      view.destroy()
      viewRef.current = null
    }
  }, [])

  // Sync external value changes (e.g. loading a different document).
  useEffect(() => {
    const view = viewRef.current
    if (!view) return
    const current = view.state.doc.toString()
    if (current === value) return
    view.dispatch({
      changes: { from: 0, to: current.length, insert: value },
    })
    valueRef.current = value
  }, [value])

  // When highlightQuery changes, force a re-render of decorations and scroll to first hit.
  useEffect(() => {
    const view = viewRef.current
    if (!view) return
    // Trigger a view update so the plugin re-computes decorations.
    view.dispatch({})
    // Scroll to first match.
    const q = highlightQuery.trim().toLowerCase()
    if (!q) return
    const text = view.state.doc.toString().toLowerCase()
    const idx = text.indexOf(q)
    if (idx >= 0) {
      view.dispatch({ effects: EditorView.scrollIntoView(idx, { y: 'center' }) })
    }
  }, [highlightQuery])

  return <div ref={containerRef} className="h-full w-full overflow-hidden" />
})
