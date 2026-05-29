import { useEffect, useRef } from 'react'
import { EditorView, Decoration, keymap, lineNumbers, highlightActiveLine, ViewPlugin, type DecorationSet } from '@codemirror/view'
import { EditorState, RangeSetBuilder } from '@codemirror/state'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language'

interface MarkdownEditorProps {
  value: string
  onChange: (value: string) => void
  onSave?: () => void
  highlightQuery?: string
}

const lightTheme = EditorView.theme({
  '&': {
    height: '100%',
    backgroundColor: '#ffffff',
  },
  '.cm-content': {
    padding: '12px 16px',
    fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
    fontSize: '14px',
    lineHeight: '1.6',
  },
  '.cm-gutters': {
    backgroundColor: '#f8f9fa',
    borderRight: '1px solid #e9ecef',
    color: '#adb5bd',
  },
  '.cm-activeLineGutter': {
    backgroundColor: '#e9ecef',
  },
  '.cm-activeLine': {
    backgroundColor: '#f8f9fa',
  },
  '.cm-cursor': {
    borderLeftColor: '#1a1a1a',
  },
  '.cm-selectionBackground': {
    backgroundColor: '#dbeafe !important',
  },
  '.cm-search-highlight': {
    backgroundColor: '#fef08a',
    borderRadius: '2px',
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

export function MarkdownEditor({ value, onChange, onSave, highlightQuery = '' }: MarkdownEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const valueRef = useRef(value)
  const onChangeRef = useRef(onChange)
  const onSaveRef = useRef(onSave)
  const queryRef = useRef(highlightQuery)

  onChangeRef.current = onChange
  onSaveRef.current = onSave
  queryRef.current = highlightQuery

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
        lightTheme,
        keymap.of([...defaultKeymap, ...historyKeymap]),
        saveKeymap,
        highlightPlugin,
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            const newValue = update.state.doc.toString()
            valueRef.current = newValue
            onChangeRef.current(newValue)
          }
        }),
        EditorView.lineWrapping,
      ],
    })

    const view = new EditorView({ state, parent: containerRef.current })
    viewRef.current = view

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
}
