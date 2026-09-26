INSERT INTO users (name, email) VALUES
    ('Иван Иванов', 'ivan@example.com'),
    ('Мария Смирнова', 'maria@example.com'),
    ('Сергей Кузнецов (IT-отдел)', 'it@example.com');

INSERT INTO categories (name) VALUES
    ('Оборудование'),
    ('Программное обеспечение'),
    ('Сеть и интернет');

INSERT INTO tickets (title, description, status, user_id, category_id) VALUES
    ('Не работает проектор', 'Проектор в аудитории 305 не включается', 'new', 1, 1),
    ('Нет Wi-Fi в библиотеке', 'Сеть eduroam не подключается', 'in_progress', 2, 3),
    ('Установить MATLAB', 'Нужен MATLAB в компьютерном классе 210', 'done', 1, 2),
    ('Сломалась мышь', NULL, 'new', 2, 1);

INSERT INTO comments (ticket_id, user_id, text) VALUES
    (1, 3, 'Заявка принята, подойдем после обеда'),
    (2, 3, 'Проблема на точке доступа, меняем'),
    (2, 2, 'Спасибо!');

INSERT INTO attachments (ticket_id, file_name, file_url) VALUES
    (1, 'projector.jpg', '/uploads/projector.jpg');
