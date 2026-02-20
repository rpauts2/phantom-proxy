'use client';

import { type FC, type ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface TableProps {
  children: ReactNode;
  className?: string;
}

export const Table: FC<TableProps> = ({ children, className }) => {
  return (
    <div className={cn('overflow-x-auto', className)}>
      <table className="w-full text-sm">{children}</table>
    </div>
  );
};

interface TableHeaderProps {
  children: ReactNode;
  className?: string;
}

export const TableHeader: FC<TableHeaderProps> = ({ children, className }) => {
  return (
    <thead>
      <tr className={cn('text-left text-zinc-400 border-b border-zinc-800', className)}>
        {children}
      </tr>
    </thead>
  );
};

interface TableBodyProps {
  children: ReactNode;
  className?: string;
}

export const TableBody: FC<TableBodyProps> = ({ children, className }) => {
  return <tbody className={cn('', className)}>{children}</tbody>;
};

interface TableRowProps {
  children: ReactNode;
  className?: string;
}

export const TableRow: FC<TableRowProps> = ({ children, className }) => {
  return <tr className={cn('border-b border-zinc-800/50', className)}>{children}</tr>;
};

interface TableCellProps {
  children: ReactNode;
  className?: string;
}

export const TableCell: FC<TableCellProps> = ({ children, className }) => {
  return <td className={cn('p-2', className)}>{children}</td>;
};

interface TableHeadProps {
  children: ReactNode;
  className?: string;
}

export const TableHead: FC<TableHeadProps> = ({ children, className }) => {
  return <th className={cn('p-2 font-semibold', className)}>{children}</th>;
};
