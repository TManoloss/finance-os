"use client";

import React, { useMemo } from 'react';
import { Treemap, Tooltip, ResponsiveContainer } from 'recharts';
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { formatCurrency } from '@/lib/utils';
import { Badge } from "@/components/ui/badge";

interface MerchantData {
  merchant_name: string;
  total_amount: number;
  tx_count: number;
  dependency_level: string;
}

interface CategoryData {
  category_name: string;
  total_amount: number;
  merchants: MerchantData[];
}

interface DependencyMapResponse {
  dependency_map: CategoryData[];
  insight: string;
}

const COLORS = ['#7C6FFF', '#4ECDC4', '#FF6B6B', '#FFD93D', '#A569BD', '#38E54D', '#FC5C9C'];

const CustomizedContent = (props: any) => {
  const { root, depth, x, y, width, height, index, payload, colors, rank, name } = props;

  return (
    <g>
      <rect
        x={x}
        y={y}
        width={width}
        height={height}
        style={{
          fill: depth < 2 ? colors[Math.floor((index / root.children.length) * 6)] : '#ffffff00',
          stroke: '#2A2A3A',
          strokeWidth: 2 / (depth + 1e-10),
          strokeOpacity: 1 / (depth + 1e-10),
        }}
      />
      {
        depth === 1 ? (
          <text
            x={x + width / 2}
            y={y + height / 2 + 7}
            textAnchor="middle"
            fill="#fff"
            fontSize={14}
            fontWeight="bold"
          >
            {name}
          </text>
        ) : null
      }
      {
        depth === 2 && width > 50 && height > 30 ? (
          <text
            x={x + 4}
            y={y + 18}
            fill="#fff"
            fontSize={12}
            fillOpacity={0.9}
          >
            {name}
          </text>
        ) : null
      }
    </g>
  );
};

export function DependencyTreemap({ data }: { data: DependencyMapResponse }) {
  const formattedData = useMemo(() => {
    if (!data || !data.dependency_map) return [];
    return data.dependency_map.map(cat => ({
      name: cat.category_name,
      children: cat.merchants.map(m => ({
        name: m.merchant_name,
        size: m.total_amount,
        tx_count: m.tx_count,
        dependency_level: m.dependency_level
      }))
    }));
  }, [data]);

  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      return (
        <div className="bg-[#1A1A24] p-3 border border-[#2A2A3A] rounded-lg shadow-xl">
          <p className="font-bold text-[#F0F0F5] mb-1">{data.name}</p>
          <p className="text-sm text-[#8888A0]">Valor: {formatCurrency(data.size)}</p>
          <p className="text-sm text-[#8888A0]">Transações: {data.tx_count}</p>
          <div className="mt-2">
            <Badge variant="outline" className={`
              ${data.dependency_level === 'Crítica' ? 'text-[#FF6B6B] border-[#FF6B6B]' : 
                data.dependency_level === 'Alta' ? 'text-[#FFD93D] border-[#FFD93D]' : 
                'text-[#4ECDC4] border-[#4ECDC4]'}
            `}>
              Dependência {data.dependency_level}
            </Badge>
          </div>
        </div>
      );
    }
    return null;
  };

  if (!data || !data.dependency_map) return null;

  return (
    <Card className="bg-[#111118] border-[#2A2A3A]">
      <CardHeader>
        <CardTitle className="text-[#F0F0F5]">Mapa de Dependência</CardTitle>
        <p className="text-sm text-[#8888A0]">{data.insight}</p>
      </CardHeader>
      <CardContent>
        <div className="h-[400px] w-full">
          <ResponsiveContainer width="100%" height="100%">
            <Treemap
              data={formattedData}
              dataKey="size"
              aspectRatio={4 / 3}
              stroke="#fff"
              fill="#8884d8"
              content={<CustomizedContent colors={COLORS} />}
            >
              <Tooltip content={<CustomTooltip />} />
            </Treemap>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}