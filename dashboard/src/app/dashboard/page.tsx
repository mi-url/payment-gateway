"use client";

import { useEffect, useState, useCallback } from "react";
import {
  ArrowUpRight,
  ArrowDownRight,
  DollarSign,
  Activity,
  CheckCircle2,
  XCircle,
  Clock,
  ArrowRight,
  CreditCard,
  Calendar,
} from "lucide-react";
import { createClient } from "@/lib/supabase/client";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Area,
  AreaChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  CartesianGrid,
} from "recharts";
import Link from "next/link";
import { Skeleton } from "@/components/ui/skeleton";

interface Transaction {
  id: string;
  amount: number;
  status: string;
  bank_code: string;
  payer_phone: string;
  initiated_at: string;
  idempotency_key: string;
}

const statusConfig = {
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

function timeAgo(dateStr: string): string {
  const now = new Date();
  const date = new Date(dateStr);
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes} min ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

function formatBs(amount: number): string {
  return `Bs. ${amount.toLocaleString("es-VE", { minimumFractionDigits: 2 })}`;
}

export default function DashboardPage() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [chartData, setChartData] = useState<any[]>([]);
  const [timeRange, setTimeRange] = useState("7d");

  const fetchData = useCallback(async () => {
    setLoading(true);
    const supabase = createClient();
    
    // Determine start date based on timeRange
    const startDate = new Date();
    startDate.setHours(0, 0, 0, 0);
    
    let daysToFetch = 7;
    if (timeRange === "7d") {
      startDate.setDate(startDate.getDate() - 6); // Last 7 days including today
      daysToFetch = 7;
    } else if (timeRange === "30d") {
      startDate.setDate(startDate.getDate() - 29); // Last 30 days
      daysToFetch = 30;
    } else if (timeRange === "month") {
      startDate.setDate(1); // 1st of current month
      const now = new Date();
      daysToFetch = now.getDate();
    }

    const { data, error } = await supabase
      .from("transactions")
      .select("id, amount, status, bank_code, payer_phone, initiated_at, idempotency_key")
      .gte("initiated_at", startDate.toISOString())
      .order("initiated_at", { ascending: false });

    if (!error && data) {
      setTransactions(data);
      
      // Process Chart Data
      const groupedData: Record<string, number> = {};
      
      // Initialize all days in range with 0 to ensure continuous chart
      for (let i = 0; i < daysToFetch; i++) {
        const d = new Date(startDate);
        d.setDate(d.getDate() + i);
        
        let label = "";
        if (daysToFetch <= 7) {
          label = d.toLocaleDateString("en-US", { weekday: "short" });
        } else {
          label = d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
        }
        groupedData[label] = 0;
      }

      // Sum amounts for SUCCESS transactions
      data.forEach(txn => {
        if (txn.status === "SUCCESS") {
          const d = new Date(txn.initiated_at);
          let label = "";
          if (daysToFetch <= 7) {
            label = d.toLocaleDateString("en-US", { weekday: "short" });
          } else {
            label = d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
          }
          if (groupedData[label] !== undefined) {
            groupedData[label] += txn.amount;
          }
        }
      });

      const newChartData = Object.keys(groupedData).map(key => ({
        date: key,
        revenue: groupedData[key]
      }));

      setChartData(newChartData);
    }
    setLoading(false);
  }, [timeRange]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Derived Metrics
  const totalRevenue = transactions
    .filter((t) => t.status === "SUCCESS")
    .reduce((sum, t) => sum + t.amount, 0);
    
  const totalCount = transactions.length;
  const successCount = transactions.filter((t) => t.status === "SUCCESS").length;
  const successRate = totalCount > 0 ? ((successCount / totalCount) * 100).toFixed(1) : "0.0";
  const pendingCount = transactions.filter(
    (t) => t.status === "PENDING_RECONCILIATION" || t.status === "PROCESSING"
  ).length;

  const stats = [
    {
      label: "Total Revenue",
      value: formatBs(totalRevenue),
      change: successCount > 0 ? `${successCount} txns` : "—",
      trend: "up" as const,
      icon: DollarSign,
    },
    {
      label: "Transactions",
      value: String(totalCount),
      change: "in period",
      trend: "up" as const,
      icon: Activity,
    },
    {
      label: "Success Rate",
      value: `${successRate}%`,
      change: `${successCount}/${totalCount}`,
      trend: Number(successRate) >= 90 ? ("up" as const) : ("down" as const),
      icon: CheckCircle2,
    },
    {
      label: "Pending",
      value: String(pendingCount),
      change: pendingCount === 0 ? "all clear" : "needs review",
      trend: pendingCount === 0 ? ("up" as const) : ("down" as const),
      icon: Clock,
    },
  ];

  return (
    <div className="max-w-6xl space-y-8 animate-in fade-in duration-500">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Overview</h1>
          <p className="text-muted-foreground mt-1">
            {loading ? "Loading data from Supabase..." : "Live metrics from your payment gateway."}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center bg-card/50 backdrop-blur-sm border border-border/50 rounded-md px-1 py-1">
            <Calendar className="w-4 h-4 text-muted-foreground ml-2 mr-1" />
            <Select value={timeRange} onValueChange={setTimeRange} disabled={loading}>
              <SelectTrigger className="w-[130px] border-0 bg-transparent h-8 shadow-none focus:ring-0">
                <SelectValue placeholder="Select Range" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="7d">Last 7 Days</SelectItem>
                <SelectItem value="30d">Last 30 Days</SelectItem>
                <SelectItem value="month">This Month</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button size="sm" className="h-10 ml-2">Create Payment Link</Button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat, i) => (
          <Card key={i} className="bg-card/50 backdrop-blur-sm shadow-sm border-border/50">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {stat.label}
              </CardTitle>
              <div className="p-2 bg-secondary/50 rounded-md">
                <stat.icon className="h-4 w-4 text-primary" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold tracking-tight">
                {loading ? <Skeleton className="h-8 w-24" /> : stat.value}
              </div>
              <p className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
                <span
                  className={`inline-flex items-center gap-0.5 font-medium ${
                    stat.trend === "up" ? "text-emerald-500" : "text-rose-500"
                  }`}
                >
                  {stat.trend === "up" ? (
                    <ArrowUpRight className="h-3 w-3" />
                  ) : (
                    <ArrowDownRight className="h-3 w-3" />
                  )}
                  {stat.change}
                </span>
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Chart */}
        <Card className="lg:col-span-2 bg-card/50 backdrop-blur-sm shadow-sm border-border/50 flex flex-col">
          <CardHeader>
            <CardTitle>Revenue Insights</CardTitle>
            <CardDescription>
              Volume processed successfully based on your selected range.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-1">
            <div className="h-[300px] w-full mt-4">
              {loading ? (
                <Skeleton className="h-full w-full rounded-xl" />
              ) : chartData.length === 0 || totalRevenue === 0 ? (
                 <div className="h-full w-full flex flex-col items-center justify-center text-muted-foreground border border-dashed border-border/50 rounded-xl">
                   <Activity className="h-8 w-8 mb-2 opacity-20" />
                   <p className="text-sm">No revenue data available for this period.</p>
                 </div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                    <defs>
                      <linearGradient id="colorRevenue" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="var(--primary)" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="var(--primary)" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border)" opacity={0.5} />
                    <XAxis
                      dataKey="date"
                      axisLine={false}
                      tickLine={false}
                      tick={{ fill: 'var(--muted-foreground)', fontSize: 12 }}
                      dy={10}
                    />
                    <YAxis
                      axisLine={false}
                      tickLine={false}
                      tick={{ fill: 'var(--muted-foreground)', fontSize: 12 }}
                      tickFormatter={(value) => `${value >= 1000 ? (value / 1000).toFixed(1) + 'k' : value}`}
                      dx={-10}
                    />
                    <Tooltip
                      contentStyle={{ backgroundColor: 'var(--card)', borderColor: 'var(--border)', borderRadius: '8px' }}
                      itemStyle={{ color: 'var(--foreground)', fontWeight: 'bold' }}
                      formatter={(value: number) => [formatBs(value), "Revenue"]}
                    />
                    <Area
                      type="monotone"
                      dataKey="revenue"
                      stroke="var(--primary)"
                      strokeWidth={2}
                      fillOpacity={1}
                      fill="url(#colorRevenue)"
                    />
                  </AreaChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Recent Transactions */}
        <Card className="bg-card/50 backdrop-blur-sm shadow-sm border-border/50 flex flex-col">
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle>Recent Activity</CardTitle>
              <CardDescription>Latest transactions in period</CardDescription>
            </div>
            <Link href="/dashboard/transactions" className={buttonVariants({ variant: "ghost", size: "icon" })}>
              <ArrowRight className="h-4 w-4" />
            </Link>
          </CardHeader>
          <CardContent className="flex-1 overflow-hidden">
            <div className="space-y-6">
              {loading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="flex items-center gap-4">
                    <Skeleton className="h-10 w-10 rounded-full" />
                    <div className="space-y-2 flex-1">
                      <Skeleton className="h-4 w-24" />
                      <Skeleton className="h-3 w-32" />
                    </div>
                  </div>
                ))
              ) : transactions.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground flex flex-col items-center">
                  <CreditCard className="h-8 w-8 mb-3 opacity-20" />
                  <p className="text-sm">No transactions yet.</p>
                </div>
              ) : (
                transactions.slice(0, 10).map((txn) => {
                  const config = statusConfig[txn.status as keyof typeof statusConfig] || statusConfig.PROCESSING;
                  const StatusIcon = config.icon;
                  return (
                    <div key={txn.id} className="flex items-center justify-between group cursor-pointer">
                      <div className="flex items-center gap-3">
                        <div className={`flex h-9 w-9 items-center justify-center rounded-full border ${config.bg} ${config.border}`}>
                          <StatusIcon className={`h-4 w-4 ${config.color}`} />
                        </div>
                        <div>
                          <p className="text-sm font-medium leading-none">
                            {formatBs(txn.amount)}
                          </p>
                          <p className="text-[11px] text-muted-foreground mt-1">
                            {bankNames[txn.bank_code] || txn.bank_code} • {timeAgo(txn.initiated_at)}
                          </p>
                        </div>
                      </div>
                      <Badge variant="outline" className={`ml-auto font-medium text-[10px] ${config.color} ${config.bg} ${config.border}`}>
                        {txn.status.replace(/_/g, " ")}
                      </Badge>
                    </div>
                  );
                })
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
