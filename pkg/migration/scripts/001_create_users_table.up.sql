CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Уникальный идентификатор пользователя
                       email VARCHAR(255) UNIQUE NOT NULL,            -- Email, уникальный и обязательный
                       username VARCHAR(50) UNIQUE NOT NULL,          -- Публичное имя
                       password_hash TEXT NOT NULL,                   -- Хеш пароля (никогда не храним в открытом виде)
                       first_name VARCHAR(100),                       -- Имя пользователя
                       last_name VARCHAR(100),                        -- Фамилия
                       is_active BOOLEAN DEFAULT TRUE,                -- Активен ли пользователь
                       role VARCHAR(50) DEFAULT 'user',               -- Роль (user/admin и т.д.)
                       last_login_at TIMESTAMPTZ,                     -- Время последнего входа
                       created_at TIMESTAMPTZ DEFAULT NOW(),          -- Дата регистрации
                       updated_at TIMESTAMPTZ DEFAULT NOW()           -- Последнее обновление
);
