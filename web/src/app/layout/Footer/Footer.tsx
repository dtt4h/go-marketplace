import { Link } from 'react-router-dom'

import cls from './Footer.module.scss'

export function Footer() {
  return (
    <footer className={cls.footer}>
      <div className={cls.footerTop}>
        <div className={cls.logoContainer}>
          <Link
            className={cls.logo}
            to="/"
            aria-label="На главную"
          >
            <span className={cls.logoLetter}>T</span>
            <p className={cls.logoWord}>Тарелка</p>
          </Link>
          <p className={cls.descLogo}>
            Маркетплейс антиквариата с экспертизой
            подлинности и защитой сделки. Покупайте и
            продавайте редкости безопасно.
          </p>
        </div>
        <nav
          className={cls.navColumns}
          aria-labelledby="footer-buyers"
        >
          <h2 className={cls.columnName} id="footer-buyers">
            Покупателям
          </h2>
          <ul className={cls.navList}>
            <li>
              <Link to="/catalog">
                Каталог
              </Link>
            </li>
            <li>
              <Link to="/cart">
                Корзина
              </Link>
            </li>
            <li>
              <Link to="/guarantees">
                Гарантии и возврат
              </Link>
            </li>
          </ul>
        </nav>

        <nav
          className={cls.navColumns}
          aria-labelledby="footer-sellers"
        >
          <h2 className={cls.columnName} id="footer-sellers">
            Продавцам
          </h2>
          <ul className={cls.navList}>
            <li>
              <Link to="/seller">
                Стать продавцом
              </Link>
            </li>
            <li>
              <Link to="/seller/dashboard">
                Панель магазина
              </Link>
            </li>
            <li>
              <Link to="/seller/publication-rules">
                Правила публикации
              </Link>
            </li>
            <li>
              <Link to="/seller/fees">
                Комиссии
              </Link>
            </li>
          </ul>
        </nav>

        <nav
          className={cls.navColumns}
          aria-labelledby="footer-contacts"
        >
          <h2 className={cls.columnName} id="footer-contacts">
            Контакты
          </h2>
          <ul className={cls.navList}>
            <li>
              <a href="tel:+74950000000">
                +7 495 000-00-00
              </a>
            </li>
            <li>
              <a href="mailto:hello@tarelka.ru">
                hello@tarelka.ru
              </a>
            </li>
            <li>
              <address className={cls.contactAddress}>
                Москва, Столешников пер., 5
              </address>
            </li>
          </ul>
        </nav>
      </div>
      <div className={cls.footerBot}>
        <div className={cls.footerBotContainer}>
          <p className={cls.footerDesc}>
            © 2026 Тарелка. Демонстрационный прототип.
          </p>
          <div className={cls.footerPolicy}>
            <Link to="/offer">Публичная оферта</Link>
            <span aria-hidden="true"> · </span>
            <Link to="/privacy">Политика конфиденциальности</Link>
          </div>
        </div>
      </div>
    </footer>
  )
}
