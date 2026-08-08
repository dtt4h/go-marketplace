import type { ProductDetails } from '../type'
import { mockProducts } from './product'

const descriptions = [
  'Редкий предмет в хорошем состоянии. Перед публикацией товар прошёл визуальную проверку.',
  'Коллекционный предмет с выразительными деталями и аккуратной отделкой.',
  'Практичный предмет для интерьера, изготовленный из качественных материалов.',
  'Винтажная вещь с естественными следами времени и хорошо сохранившимися элементами.',
  'Оригинальный предмет, который станет заметной частью домашней коллекции.',
]

export const mockProductDetails: ProductDetails[] =
  mockProducts.map((product, index) => ({
    ...product,
    description:
      descriptions[index % descriptions.length],
    status: 'active',
    images: product.images
      ? Array.from({ length: 4 }, (_, imageIndex) => ({
          id: product.id * 10 + imageIndex,
          url:
            `https://picsum.photos/seed/` +
            `product-${product.id}-${imageIndex}/900/900`,
          position: imageIndex,
        }))
      : undefined,
    updated_at: product.created_at,
  }))
