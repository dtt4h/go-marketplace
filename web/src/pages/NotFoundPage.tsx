import { Link } from 'react-router-dom'

export function NotFoundPage() {
  return (
    <main>
      <p>Страница не найдена</p>
      <Link to="/">Вернуться в каталог</Link>
    </main>
  )
}