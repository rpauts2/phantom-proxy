'use client';

import { type FC } from 'react';
import { cn } from '@/lib/utils';

interface CardProps {
  title: string;
  value: number | string;
  loading?: boolean;
  icon?: FC<{ className?: string }>;
  className?: string;
}

export const Card: FC<CardProps> = ({ title, value, loading = false, icon: Icon, className }) => {
  return (
    <div className={cn('bg-zinc-900 rounded-lg p-4 border border-zinc-800', className)}>
      <div className="flex items-center justify-between">
        <p className="text-zinc-400 text-sm">{title}</p>
        {Icon && <Icon className="w-5 h-5 text-zinc-500" />}
      </div>
      <p className="text-2xl font-bold mt-2 text-white">
        {loading ? (
          <span className="animate-pulse">...</span>
        ) : (
          value
        )}
      </p>
    </div>
  );
};
