import type { ProductListItem } from "../../type"

type ProductCardProps = {
  product: ProductListItem
}

export function ProductCard({
  product,
}: ProductCardProps) {
  const previewImage = product.images?.[0]

  return (
    <article>
      {previewImage ? (
        <img
          src={previewImage.url}
          alt={product.title}
        />
      ) : (
        <div>Нет изображения</div>
      )}

      <h2>{product.title}</h2>
      
      <p>{product.price} ₽</p>

      <p>
        {product.store?.name ?? 'Магазин не указан'}
      </p>

      {product.stock > 0 ? (
        <span>В наличии</span>
      ) : (
        <span>Нет в наличии</span>
      )}
    </article>
  )
}