"use client";

import { Button } from "@/components/ui/button";
import { CaretLeft, CaretRight } from "@phosphor-icons/react";

interface PaginationProps {
  page: number;
  total: number;
  limit: number;
  onPageChange: (page: number) => void;
}

export function Pagination({ page, total, limit, onPageChange }: PaginationProps) {
  const totalPages = Math.max(1, Math.ceil(total / limit));
  if (totalPages <= 1) return null;

  return (
    <div className="flex items-center justify-between border-t border-border-subtle pt-4">
      <p className="text-sm text-muted">
        Page {page} of {totalPages} ({total} items)
      </p>
      <div className="flex gap-2">
        <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => onPageChange(page - 1)}>
          <CaretLeft size={14} weight="bold" />
          Prev
        </Button>
        <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => onPageChange(page + 1)}>
          Next
          <CaretRight size={14} weight="bold" />
        </Button>
      </div>
    </div>
  );
}
