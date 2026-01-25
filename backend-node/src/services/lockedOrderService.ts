import { PrismaClient } from "@prisma/client";

interface LockedOrderItem {
  sku: string;
  productName: string;
  variationName?: string;
  qty: number;
}

export class LockedOrderService {
  private prisma: PrismaClient;

  constructor(prisma: PrismaClient) {
    this.prisma = prisma;
  }

  async saveLockedOrders(
    tenantId: string,
    items: LockedOrderItem[]
  ): Promise<number> {
    // Delete existing locked orders for this tenant
    await this.prisma.lockedOrder.deleteMany({
      where: { tenantId },
    });

    if (items.length === 0) return 0;

    // Bulk create new locked orders
    const createData = items.map((item) => ({
      tenantId,
      sku: item.sku || "",
      productName: item.productName || "",
      variationName: item.variationName || null,
      qty: item.qty || 0,
    }));

    await this.prisma.lockedOrder.createMany({
      data: createData,
    });

    return items.length;
  }

  async getLockedOrders(tenantId: string): Promise<LockedOrderItem[]> {
    const orders = await this.prisma.lockedOrder.findMany({
      where: { tenantId },
      orderBy: { qty: "desc" },
      select: {
        sku: true,
        productName: true,
        variationName: true,
        qty: true,
      },
    });

    return orders.map((order) => ({
      sku: order.sku,
      productName: order.productName,
      variationName: order.variationName || undefined,
      qty: order.qty,
    }));
  }

  async getTotalQty(tenantId: string): Promise<number> {
    const result = await this.prisma.lockedOrder.aggregate({
      where: { tenantId },
      _sum: { qty: true },
    });

    return result._sum.qty || 0;
  }
}
