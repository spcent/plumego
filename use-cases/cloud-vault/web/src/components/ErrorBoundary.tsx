import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
  errorInfo?: ErrorInfo
}

export default class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
    this.setState({ errorInfo })
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-background p-8">
          <div className="max-w-2xl w-full space-y-6">
            <div className="text-center space-y-2">
              <h1 className="text-2xl font-semibold text-destructive">Something went wrong</h1>
              <p className="text-sm text-muted-foreground">
                The application encountered an unexpected error.
              </p>
            </div>

            {import.meta.env.DEV && this.state.error && (
              <div className="border border-destructive/30 rounded-lg p-4 bg-destructive/5 space-y-2">
                <h2 className="text-sm font-medium text-foreground">Error Details</h2>
                <pre className="text-xs text-destructive overflow-auto max-h-40 bg-background p-2 rounded">
                  {this.state.error.toString()}
                </pre>
                {this.state.errorInfo?.componentStack && (
                  <pre className="text-xs text-muted-foreground overflow-auto max-h-40 bg-background p-2 rounded">
                    {this.state.errorInfo.componentStack}
                  </pre>
                )}
              </div>
            )}

            <div className="flex gap-3 justify-center">
              <button
                onClick={() => window.location.reload()}
                className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded hover:bg-primary/90"
              >
                Reload Application
              </button>
            </div>

            <p className="text-xs text-center text-muted-foreground">
              If the problem persists, try clearing your browser cache or contact support.
            </p>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
