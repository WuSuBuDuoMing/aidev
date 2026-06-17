import React, { useState, useEffect, useRef } from 'react'
import { SendMessage, GetConfig, GetStatus, ListSessions, GetContextUsage, NewConversation, UpdateConfig, GetProviders } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime'

function App() {
  const [config, setConfig] = useState({})
  const [status, setStatus] = useState({})
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [sessions, setSessions] = useState([])
  const [contextUsage, setContextUsage] = useState({})
  const [providers, setProviders] = useState([])
  const [streamText, setStreamText] = useState('')
  const [showSettings, setShowSettings] = useState(false)
  const messagesEndRef = useRef(null)

  useEffect(() => {
    loadData()
    // Listen for streaming tokens
    const tokenUnsub = EventsOn('stream:token', (token) => {
      setStreamText(prev => prev + token)
    })
    const toolStartUnsub = EventsOn('tool:start', (data) => {
      setMessages(prev => [...prev, {
        role: 'tool',
        content: `🔧 ${data.name}`,
        type: 'tool-start'
      }])
    })
    const toolEndUnsub = EventsOn('tool:end', (data) => {
      setMessages(prev => [...prev, {
        role: 'tool',
        content: data.success ? `✓ ${data.name}` : `✗ ${data.name}: ${data.error}`,
        type: data.success ? 'tool-success' : 'tool-error'
      }])
    })
    return () => {
      tokenUnsub()
      toolStartUnsub()
      toolEndUnsub()
    }
  }, [])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, streamText])

  async function loadData() {
    const [cfg, st, sess, ctx] = await Promise.all([
      GetConfig(), GetStatus(), ListSessions(), GetContextUsage()
    ])
    setConfig(cfg)
    setStatus(st)
    setSessions(sess || [])
    setContextUsage(ctx)
    const provs = await GetProviders()
    setProviders(provs || [])
  }

  async function handleSend() {
    if (!input.trim() || loading) return
    const userMsg = input.trim()
    setInput('')
    setMessages(prev => [...prev, { role: 'user', content: userMsg }])
    setLoading(true)
    setStreamText('')

    try {
      const result = await SendMessage(userMsg)
      if (result.error) {
        setMessages(prev => [...prev, { role: 'assistant', content: `Error: ${result.error}` }])
      } else {
        if (streamText) {
          setStreamText('')
        }
        setMessages(prev => [...prev, {
          role: 'assistant',
          content: result.content,
          tokens: result.tokens,
          cost: result.cost,
        }])
      }
      await loadData()
    } catch (err) {
      setMessages(prev => [...prev, { role: 'assistant', content: `Error: ${err}` }])
    }
    setLoading(false)
    setStreamText('')
  }

  async function handleNewChat() {
    await NewConversation()
    setMessages([])
    setStreamText('')
    await loadData()
  }

  const percentage = parseFloat(contextUsage?.percentage) || 0
  const barColor = percentage > 80 ? 'bg-neocode-error' : percentage > 60 ? 'bg-neocode-warning' : 'bg-neocode-success'

  return (
    <div className="flex h-screen bg-neocode-bg">
      {/* Sidebar */}
      <div className="w-64 bg-neocode-surface border-r border-neocode-border flex flex-col">
        <div className="p-4 border-b border-neocode-border">
          <h1 className="text-lg font-bold text-neocode-primary">NeoCode</h1>
          <p className="text-xs text-neocode-muted">{config.provider} / {config.model}</p>
        </div>
        <div className="p-3">
          <button onClick={handleNewChat}
            className="w-full px-3 py-2 bg-neocode-primary text-white rounded-md text-sm hover:opacity-80 transition">
            + New Chat
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-2">
          <p className="text-xs text-neocode-muted px-2 mb-2">Recent Sessions</p>
          {sessions?.map(s => (
            <div key={s.id} className="px-2 py-1.5 rounded text-sm text-neocode-text hover:bg-neocode-border cursor-pointer truncate">
              {s.title}
            </div>
          ))}
        </div>
        {/* Context Usage */}
        <div className="p-3 border-t border-neocode-border">
          <p className="text-xs text-neocode-muted mb-1">Context Usage</p>
          <div className="w-full bg-neocode-border rounded-full h-2">
            <div className={`${barColor} h-2 rounded-full transition-all`} style={{ width: `${Math.min(percentage, 100)}%` }} />
          </div>
          <p className="text-xs text-neocode-muted mt-1">
            {contextUsage?.currentTokens?.toLocaleString()} / {contextUsage?.contextWindow?.toLocaleString()} tokens ({contextUsage?.percentage})
          </p>
        </div>
      </div>

      {/* Main Area */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <div className="h-12 bg-neocode-surface border-b border-neocode-border flex items-center px-4 justify-between">
          <div className="flex items-center gap-3">
            <span className="text-sm text-neocode-text">{config.model}</span>
            <span className="text-xs px-2 py-0.5 rounded bg-neocode-border text-neocode-muted">{config.effort}</span>
            <span className="text-xs px-2 py-0.5 rounded bg-neocode-border text-neocode-muted">{config.mode}</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-neocode-muted">{status.toolsCount} tools</span>
            <button onClick={() => setShowSettings(!showSettings)}
              className="text-xs px-2 py-1 rounded hover:bg-neocode-border text-neocode-muted">
              ⚙ Settings
            </button>
          </div>
        </div>

        {/* Settings Panel */}
        {showSettings && (
          <div className="bg-neocode-surface border-b border-neocode-border p-4">
            <div className="grid grid-cols-3 gap-4 max-w-2xl">
              <div>
                <label className="text-xs text-neocode-muted block mb-1">Provider</label>
                <select value={config.provider} onChange={e => UpdateConfig({ provider: e.target.value }).then(setConfig)}
                  className="w-full bg-neocode-bg border border-neocode-border rounded px-2 py-1 text-sm text-neocode-text">
                  {providers.map(p => <option key={p.name} value={p.name}>{p.name}</option>)}
                </select>
              </div>
              <div>
                <label className="text-xs text-neocode-muted block mb-1">Effort</label>
                <select value={config.effort} onChange={e => UpdateConfig({ effort: e.target.value }).then(setConfig)}
                  className="w-full bg-neocode-bg border border-neocode-border rounded px-2 py-1 text-sm text-neocode-text">
                  {['low', 'medium', 'high', 'xhigh', 'ultracode'].map(e => <option key={e} value={e}>{e}</option>)}
                </select>
              </div>
              <div>
                <label className="text-xs text-neocode-muted block mb-1">Mode</label>
                <select value={config.mode} onChange={e => UpdateConfig({ mode: e.target.value }).then(setConfig)}
                  className="w-full bg-neocode-bg border border-neocode-border rounded px-2 py-1 text-sm text-neocode-text">
                  {['ask', 'auto', 'plan', 'edit'].map(m => <option key={m} value={m}>{m}</option>)}
                </select>
              </div>
            </div>
          </div>
        )}

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.length === 0 && (
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <h2 className="text-2xl font-bold text-neocode-primary mb-2">NeoCode</h2>
                <p className="text-neocode-muted text-sm">AI coding agent for Chinese and international models</p>
                <p className="text-neocode-muted text-xs mt-1">Type a message to start coding</p>
              </div>
            </div>
          )}
          {messages.map((msg, i) => (
            <div key={i} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
              <div className={`max-w-[80%] rounded-lg px-4 py-2 ${
                msg.role === 'user'
                  ? 'bg-neocode-primary text-white'
                  : msg.type?.startsWith('tool')
                    ? 'bg-neocode-surface border border-neocode-border text-neocode-muted text-xs font-mono'
                    : 'bg-neocode-surface border border-neocode-border text-neocode-text'
              }`}>
                <pre className="whitespace-pre-wrap text-sm">{msg.content}</pre>
                {msg.tokens && (
                  <p className="text-xs text-neocode-muted mt-1">
                    {msg.tokens} tokens {msg.cost > 0 && `· $${msg.cost.toFixed(4)}`}
                  </p>
                )}
              </div>
            </div>
          ))}
          {streamText && (
            <div className="flex justify-start">
              <div className="max-w-[80%] rounded-lg px-4 py-2 bg-neocode-surface border border-neocode-border text-neocode-text">
                <pre className="whitespace-pre-wrap text-sm">{streamText}</pre>
              </div>
            </div>
          )}
          {loading && !streamText && (
            <div className="flex justify-start">
              <div className="px-4 py-2 text-neocode-muted text-sm">
                <span className="animate-pulse">Thinking...</span>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Input */}
        <div className="p-4 border-t border-neocode-border">
          <div className="flex gap-2">
            <input
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && !e.shiftKey && handleSend()}
              placeholder="Type a message..."
              disabled={loading}
              className="flex-1 bg-neocode-surface border border-neocode-border rounded-lg px-4 py-2.5 text-sm text-neocode-text placeholder-neocode-muted focus:outline-none focus:border-neocode-primary disabled:opacity-50"
            />
            <button
              onClick={handleSend}
              disabled={loading || !input.trim()}
              className="px-4 py-2.5 bg-neocode-primary text-white rounded-lg text-sm hover:opacity-80 transition disabled:opacity-50"
            >
              Send
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
