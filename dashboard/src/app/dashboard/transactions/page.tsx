"use client";

import { useEffect, useState, useCallback } from "react";
import {
  CheckCircle2,
  XCircle,
  Clock,
  Activity,
  Search,
  Filter,
  Download,
  CreditCard,
  ChevronLeft,
  ChevronRight,
  FileJson,
  Phone,
  Hash,
  Landmark,
  CalendarDays,
  Copy,
  Check,
  FilterX,
} from "lucide-react";
import { toast } from "sonner";
import { createClient } from "@/lib/supabase/client";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Separator } from "@/components/ui/separator";

interface Transaction {
  id: string;
  idempotency_key: string;
  amount: number;
  currency: string;
  status: string;
  bank_code: string;
  payer_phone: string | null;
  payer_id_document: string | null;
  bank_reference: string | null;
  error_code: string | null;
  initiated_at: string;
  raw_response?: any;
}

const statusConfig: Record<string, { color: string; bg: string; border: string; icon: React.ElementType }> = {
  SUCCESS: { color: "text-emerald-500", bg: "bg-emerald-500/10", border: "border-emerald-500/20", icon: CheckCircle2 },
  DECLINED: { color: "text-rose-500", bg: "bg-rose-500/10", border: "border-rose-500/20", icon: XCircle },
  PENDING_RECONCILIATION: { color: "text-blue-500", bg: "bg-blue-500/10", border: "border-blue-500/20", icon: Clock },
  PROCESSING: { color: "text-amber-500", bg: "bg-amber-500/10", border: "border-amber-500/20", icon: Activity },
  INITIATED: { color: "text-slate-400", bg: "bg-slate-500/10", border: "border-slate-500/20", icon: Clock },
  BANK_NETWORK_ERROR: { color: "text-rose-500", bg: "bg-rose-500/10", border: "border-rose-500/20", icon: XCircle },
};

const bankNames: Record<string, string> = {
  "0191": "BNC",
  "0105": "Mercantil",
  "0134": "Banesco",
  "0102": "Venezuela",
  "0172": "Bancamiga",
};

function CopyButton({ text, label }: { text: string; label?: string }) {
  const [copied, setCopied] = useState(false);
  const handleCopy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    toast.success(`${label || "ID"} copied to clipboard`);
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <Button variant="ghost" size="icon" className="h-6 w-6 ml-2 text-muted-foreground hover:text-foreground" onClick={handleCopy}>
      {copied ? <Check className="h-3 w-3" /> : <Copy className="h-3 w-3" />}
    </Button>
  );
}

function formatAmount(amount: number): string {
  return `Bs. ${amount.toLocaleString("es-VE", { minimumFractionDigits: 2 })}`;
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleString("es-VE", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

const PAGE_SIZE = 15;

export default function TransactionsPage() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  
  // Advanced Filters
  const [statusFilter, setStatusFilter] = useState("ALL");
  const [bankFilter, setBankFilter] = useState("ALL");
  
  // Pagination
  const [page, setPage] = useState(0);
  const [totalCount, setTotalCount] = useState(0);

  // Details Sheet
  const [selectedTxn, setSelectedTxn] = useState<Transaction | null>(null);

  // Debounce search query
  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedSearch(searchQuery);
      setPage(0); // Reset to first page on search
    }, 400);
    return () => clearTimeout(handler);
  }, [searchQuery]);

  // Reset page when filters change
  useEffect(() => {
    setPage(0);
  }, [statusFilter, bankFilter]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    const supabase = createClient();
    
    let query = supabase
      .from("transactions")
      .select("*", { count: "exact" });

    if (debouncedSearch) {
      const q = `%${debouncedSearch}%`;
      query = query.or(`idempotency_key.ilike.${q},payer_phone.ilike.${q},bank_reference.ilike.${q},status.ilike.${q}`);
    }

    if (statusFilter !== "ALL") {
      query = query.eq("status", statusFilter);
    }
    
    if (bankFilter !== "ALL") {
      query = query.eq("bank_code", bankFilter);
    }

    query = query
      .order("initiated_at", { ascending: false })
      .range(page * PAGE_SIZE, (page + 1) * PAGE_SIZE - 1);

    const { data, count, error } = await query;

    if (!error && data) {
      setTransactions(data);
      if (count !== null) setTotalCount(count);
    }
    setLoading(false);
  }, [debouncedSearch, statusFilter, bankFilter, page]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleExportCSV = () => {
    if (transactions.length === 0) return;
    
    const headers = ["ID", "Amount", "Currency", "Status", "Bank", "Payer Phone", "Reference", "Date"];
    const csvContent = [
      headers.join(","),
      ...transactions.map(t => [
        t.idempotency_key,
        t.amount,
        t.currency,
        t.status,
        bankNames[t.bank_code] || t.bank_code,
        t.payer_phone || "",
        t.bank_reference || "",
        t.initiated_at
      ].join(","))
    ].join("\\n");

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.setAttribute("href", url);
    link.setAttribute("download", `transactions_page_${page + 1}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const totalPages = Math.ceil(totalCount / PAGE_SIZE);

  return (
    <div className="max-w-6xl space-y-6 animate-in fade-in duration-500">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Transactions</h1>
          <p className="text-muted-foreground mt-1">
            {totalCount > 0 ? `${totalCount} total records` : "Loading..."}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" className="h-9" onClick={handleExportCSV}>
            <Download className="mr-2 h-4 w-4" />
            Export Page CSV
          </Button>
        </div>
      </div>

      {/* Filters Area */}
      <div className="flex flex-col sm:flex-row items-center gap-3">
        <div className="relative flex-1 w-full max-w-md">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search ID, phone, or reference..."
            className="pl-9 bg-card/50 backdrop-blur-sm border-border/50 h-9"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
        <div className="flex items-center gap-2 w-full sm:w-auto">
          <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v ?? 'all')}>
            <SelectTrigger className="w-[140px] h-9 bg-card/50 backdrop-blur-sm border-border/50">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">All Statuses</SelectItem>
              <SelectItem value="SUCCESS">Success</SelectItem>
              <SelectItem value="DECLINED">Declined</SelectItem>
              <SelectItem value="PROCESSING">Processing</SelectItem>
              <SelectItem value="PENDING_RECONCILIATION">Pending</SelectItem>
              <SelectItem value="BANK_NETWORK_ERROR">Network Error</SelectItem>
            </SelectContent>
          </Select>
          <Select value={bankFilter} onValueChange={(v) => setBankFilter(v ?? 'all')}>
            <SelectTrigger className="w-[140px] h-9 bg-card/50 backdrop-blur-sm border-border/50">
              <SelectValue placeholder="Bank" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">All Banks</SelectItem>
              <SelectItem value="0191">BNC (0191)</SelectItem>
              <SelectItem value="0105">Mercantil (0105)</SelectItem>
              <SelectItem value="0134">Banesco (0134)</SelectItem>
              <SelectItem value="0172">Bancamiga (0172)</SelectItem>
            </SelectContent>
          </Select>
          {(statusFilter !== "ALL" || bankFilter !== "ALL" || searchQuery) && (
            <Button 
              variant="ghost" 
              size="sm" 
              className="h-9 px-2 text-muted-foreground hover:text-foreground"
              onClick={() => {
                setStatusFilter("ALL");
                setBankFilter("ALL");
                setSearchQuery("");
              }}
            >
              <FilterX className="h-4 w-4 mr-1" /> Clear
            </Button>
          )}
        </div>
      </div>

      {/* Table */}
      <Card className="bg-card/50 backdrop-blur-sm border-border/50 shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow className="border-border/50 hover:bg-transparent">
                <TableHead className="w-[250px]">Transaction</TableHead>
                <TableHead>Amount</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Bank</TableHead>
                <TableHead>Reference</TableHead>
                <TableHead className="text-right">Date</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <TableRow key={i} className="border-border/50">
                    <TableCell><Skeleton className="h-4 w-32 mb-1" /><Skeleton className="h-3 w-24" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                    <TableCell><Skeleton className="h-6 w-24 rounded-full" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                    <TableCell className="text-right"><Skeleton className="h-4 w-32 ml-auto" /></TableCell>
                  </TableRow>
                ))
              ) : transactions.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="h-48 text-center">
                    <div className="flex flex-col items-center justify-center text-muted-foreground">
                      <div className="w-12 h-12 rounded-full bg-secondary flex items-center justify-center mb-3">
                        <Search className="h-6 w-6 opacity-50" />
                      </div>
                      <p className="font-medium text-foreground">No transactions found</p>
                      <p className="text-sm mt-1 max-w-sm">
                        Try adjusting your filters or search query to find what you're looking for.
                      </p>
                      {(statusFilter !== "ALL" || bankFilter !== "ALL" || searchQuery) && (
                        <Button 
                          variant="outline" 
                          size="sm" 
                          className="mt-4"
                          onClick={() => {
                            setStatusFilter("ALL");
                            setBankFilter("ALL");
                            setSearchQuery("");
                          }}
                        >
                          Clear all filters
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ) : (
                transactions.map((txn) => {
                  const config = statusConfig[txn.status] || statusConfig.PROCESSING;
                  const StatusIcon = config.icon;
                  return (
                    <TableRow 
                      key={txn.id} 
                      className="border-border/50 group hover:bg-muted/50 cursor-pointer"
                      onClick={() => setSelectedTxn(txn)}
                    >
                      <TableCell>
                        <div className="font-mono text-xs font-medium text-foreground">
                          {txn.idempotency_key}
                        </div>
                        <div className="text-[11px] text-muted-foreground mt-0.5">
                          {txn.payer_phone || "—"}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="font-semibold text-foreground">
                          {formatAmount(txn.amount)}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline" className={`font-medium ${config.color} ${config.bg} ${config.border}`}>
                          <StatusIcon className="mr-1.5 h-3 w-3" />
                          {txn.status.replace(/_/g, " ")}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-medium text-muted-foreground">
                            {bankNames[txn.bank_code] || txn.bank_code}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="font-mono text-xs text-muted-foreground">
                          {txn.bank_reference || "—"}
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="text-xs text-muted-foreground whitespace-nowrap">
                          {formatDate(txn.initiated_at)}
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
        </div>
        
        {/* Pagination Footer */}
        {!loading && totalCount > 0 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-border/50 bg-muted/20">
            <div className="text-xs text-muted-foreground">
              Showing <span className="font-medium text-foreground">{page * PAGE_SIZE + 1}</span> to <span className="font-medium text-foreground">{Math.min((page + 1) * PAGE_SIZE, totalCount)}</span> of <span className="font-medium text-foreground">{totalCount}</span>
            </div>
            <div className="flex items-center gap-2">
              <Button 
                variant="outline" 
                size="sm" 
                className="h-8 w-8 p-0" 
                onClick={() => setPage(Math.max(0, page - 1))}
                disabled={page === 0}
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
              <Button 
                variant="outline" 
                size="sm" 
                className="h-8 w-8 p-0" 
                onClick={() => setPage(page + 1)}
                disabled={page >= totalPages - 1}
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </Card>

      {/* Details Sheet */}
      <Sheet open={!!selectedTxn} onOpenChange={(open) => !open && setSelectedTxn(null)}>
        <SheetContent className="w-full sm:max-w-md md:max-w-lg overflow-y-auto bg-card border-border/50 p-0 sm:max-w-xl">
          {selectedTxn && (
            <div className="flex flex-col min-h-full">
              {/* Header */}
              <div className="px-6 py-6 border-b border-border/40">
                <SheetHeader className="pr-8">
                  <div className="flex items-center gap-2">
                    <SheetTitle className="font-mono text-lg tracking-tight">
                      {selectedTxn.idempotency_key}
                    </SheetTitle>
                    <CopyButton text={selectedTxn.idempotency_key} label="Transaction ID" />
                  </div>
                  <SheetDescription>
                    {formatDate(selectedTxn.initiated_at)}
                  </SheetDescription>
                </SheetHeader>
              </div>

              <div className="px-6 py-6 space-y-8 flex-1">
                {/* Hero Status Box - Clean & Premium */}
                <div className="relative rounded-2xl bg-zinc-950/50 border border-border/40 overflow-hidden">
                  <div className={`absolute top-0 left-0 w-full h-1 ${statusConfig[selectedTxn.status]?.bg.replace('/10', '/50')}`} />
                  <div className="p-6 flex flex-col items-center justify-center text-center">
                    <div className="mb-3">
                      {(() => {
                        const Icon = statusConfig[selectedTxn.status]?.icon || Activity;
                        return <Icon className={`w-8 h-8 ${statusConfig[selectedTxn.status]?.color}`} strokeWidth={2.5} />;
                      })()}
                    </div>
                    <h3 className="text-4xl tracking-tighter font-semibold text-foreground mb-1">
                      {formatAmount(selectedTxn.amount)}
                    </h3>
                    <div className="flex items-center gap-2">
                      <span className={`text-sm font-medium tracking-wide ${statusConfig[selectedTxn.status]?.color}`}>
                        {selectedTxn.status.replace(/_/g, " ")}
                      </span>
                      {selectedTxn.error_code && (
                        <>
                          <span className="text-muted-foreground">•</span>
                          <span className="text-sm font-medium text-destructive">
                            Code: {selectedTxn.error_code}
                          </span>
                        </>
                      )}
                    </div>
                  </div>
                </div>

                {/* Structured Data Lists */}
                <div className="space-y-6">
                  {/* Bank Details */}
                  <div className="space-y-3">
                    <h4 className="text-sm font-medium text-foreground flex items-center gap-2">
                      <Landmark className="w-4 h-4 text-primary" /> Bank Details
                    </h4>
                    <div className="rounded-xl border border-border/40 bg-card">
                      <div className="flex justify-between items-center py-3 px-4 border-b border-border/40">
                        <span className="text-sm text-muted-foreground">Bank</span>
                        <span className="text-sm font-medium">{bankNames[selectedTxn.bank_code] || selectedTxn.bank_code}</span>
                      </div>
                      <div className="flex justify-between items-center py-3 px-4">
                        <span className="text-sm text-muted-foreground">Reference</span>
                        <span className="text-sm font-mono font-medium">{selectedTxn.bank_reference || "N/A"}</span>
                      </div>
                    </div>
                  </div>

                  {/* Payer Information */}
                  <div className="space-y-3">
                    <h4 className="text-sm font-medium text-foreground flex items-center gap-2">
                      <Phone className="w-4 h-4 text-primary" /> Payer Information
                    </h4>
                    <div className="rounded-xl border border-border/40 bg-card">
                      <div className="flex justify-between items-center py-3 px-4 border-b border-border/40">
                        <span className="text-sm text-muted-foreground">Phone Number</span>
                        <span className="text-sm font-medium">{selectedTxn.payer_phone || "N/A"}</span>
                      </div>
                      <div className="flex justify-between items-center py-3 px-4">
                        <span className="text-sm text-muted-foreground">ID Document</span>
                        <span className="text-sm font-medium">{selectedTxn.payer_id_document || "N/A"}</span>
                      </div>
                    </div>
                  </div>

                  {/* System Identifiers */}
                  <div className="space-y-3">
                    <h4 className="text-sm font-medium text-foreground flex items-center gap-2">
                      <Hash className="w-4 h-4 text-primary" /> Identifiers
                    </h4>
                    <div className="rounded-xl border border-border/40 bg-card">
                      <div className="flex justify-between items-center py-3 px-4 border-b border-border/40">
                        <span className="text-sm text-muted-foreground">Internal ID</span>
                        <div className="flex items-center">
                          <span className="text-xs font-mono text-foreground/80">{selectedTxn.id}</span>
                          <CopyButton text={selectedTxn.id} label="Internal ID" />
                        </div>
                      </div>
                      <div className="flex justify-between items-center py-3 px-4">
                        <span className="text-sm text-muted-foreground">Idempotency Key</span>
                        <div className="flex items-center">
                          <span className="text-xs font-mono text-foreground/80">{selectedTxn.idempotency_key}</span>
                          <CopyButton text={selectedTxn.idempotency_key} label="Idempotency Key" />
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* System Log - Premium Terminal Look */}
                <div className="space-y-3 pt-2">
                  <h4 className="text-sm font-medium text-foreground flex items-center gap-2">
                    <FileJson className="w-4 h-4 text-primary" /> JSON Payload
                  </h4>
                  <div className="rounded-xl overflow-hidden border border-border/50 shadow-sm">
                    <div className="bg-zinc-950/80 px-4 py-2 border-b border-border/30 flex items-center justify-between">
                      <div className="flex gap-1.5">
                        <div className="w-2.5 h-2.5 rounded-full bg-rose-500/20 border border-rose-500/50" />
                        <div className="w-2.5 h-2.5 rounded-full bg-amber-500/20 border border-amber-500/50" />
                        <div className="w-2.5 h-2.5 rounded-full bg-emerald-500/20 border border-emerald-500/50" />
                      </div>
                      <span className="text-[10px] font-mono text-muted-foreground uppercase tracking-wider">transaction.processed</span>
                    </div>
                    <div className="bg-zinc-950 p-4 overflow-x-auto">
                      <pre className="text-[11px] font-mono leading-relaxed">
<span className="text-blue-400">{"{"}</span>
<br/>  <span className="text-sky-300">"event"</span><span className="text-foreground/70">:</span> <span className="text-emerald-400">"transaction.processed"</span><span className="text-foreground/70">,</span>
<br/>  <span className="text-sky-300">"timestamp"</span><span className="text-foreground/70">:</span> <span className="text-emerald-400">"{selectedTxn.initiated_at}"</span><span className="text-foreground/70">,</span>
<br/>  <span className="text-sky-300">"processor"</span><span className="text-foreground/70">:</span> <span className="text-emerald-400">"faloppa_switch_v1"</span><span className="text-foreground/70">,</span>
<br/>  <span className="text-sky-300">"bank_response"</span><span className="text-foreground/70">:</span> <span className="text-blue-400">{"{"}</span>
<br/>    <span className="text-sky-300">"code"</span><span className="text-foreground/70">:</span> <span className="text-emerald-400">"{selectedTxn.error_code || '0000'}"</span><span className="text-foreground/70">,</span>
<br/>    <span className="text-sky-300">"message"</span><span className="text-foreground/70">:</span> <span className="text-emerald-400">"{selectedTxn.status}"</span><span className="text-foreground/70">,</span>
<br/>    <span className="text-sky-300">"reference"</span><span className="text-foreground/70">:</span> <span className="text-emerald-400">"{selectedTxn.bank_reference || 'null'}"</span>
<br/>  <span className="text-blue-400">{"}"}</span>
<br/><span className="text-blue-400">{"}"}</span>
                      </pre>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </SheetContent>
      </Sheet>
    </div>
  );
}
