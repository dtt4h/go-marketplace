import type { FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../../shared/ui/Button/Button'
import { Profile } from '../../../shared/assets/icons/Profile'
import { Cart } from '../../../shared/assets/icons/Cart'
import cls from './Header.module.scss'

export function Header() {
  function handleSearchSubmit(
    event: FormEvent<HTMLFormElement>,
  ): void {
    event.preventDefault()
  }

  return (
    <header className={cls.header}>
      <div className={cls.top}>
        <p className={cls.garantText}>Проверка подлинности каждого лота экспертом · Безопасная сделка через эскроу</p>
      </div>
      <div className={cls.mid}>
        <div className={cls.logoContainer}>
          <span className={cls.logoLetter}>T</span>
          <Link to="/" aria-label="На главную" className={cls.logoWord}>Тарелка</Link>
        </div>
        <form
          className={cls.searchbarContainer}
          role="search"
          onSubmit={handleSearchSubmit}
        >
          <label
            className={cls.visuallyHidden}
            htmlFor="catalog-search"
          >
            Поиск по каталогу
          </label>
          <input 
            id="catalog-search"
            name="search"
            type="search"
            className={cls.searchInput}
            placeholder='Поиск по антиквариату: комод, монета, икона…'
          />
          <Button
            className={cls.searchButton}
            type="submit"
          >
            Найти
          </Button>
        </form>
        <div className={cls.buttonsContainer}>
          <Button
            className={cls.becomeSeller}
            type="button"
          >
            Продавать
          </Button>
          <Link
            className={cls.toProfile}
            to="/profile"
            aria-label="Открыть профиль"
            title="Открыть профиль"
          >
            <Profile aria-hidden="true" focusable="false" />
          </Link>
          <Button
            className={cls.toCart}
            type="button"
            aria-label="Открыть корзину"
            title="Открыть корзину"
          >
            <Cart aria-hidden="true" focusable="false" />
          </Button>
        </div>
      </div>
      <nav aria-label="Основная навигация" className={cls.bot}>
        <ul className={cls.categoryList}>
          <li>
            <Link to="/catalog">
              Весь каталог
            </Link>
          </li>
          <li>
            <Link to="/catalog?category_id=1">
              Мебель
            </Link>
          </li>
          <li>
            <Link to="/catalog?category_id=2">
              Живопись
            </Link>
          </li>
          <li>
            <Link to="/catalog?category_id=3">
              Монеты
            </Link>
          </li>
          <li>
            <Link to="/catalog?category_id=4">
              Часы
            </Link>
          </li>
          <li>
            <Link to="/catalog?category_id=5">
              Фарфор
            </Link>
          </li>
          <li>
            <Link to="/catalog?category_id=6">
              Освещение
            </Link>
          </li>
        </ul>
      </nav>
    </header>
  )
}
