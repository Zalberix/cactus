-- Создаёт базы для Temporal, изолируя их от app-базы 'postgres'
-- Выполняется только при первой инициализации volume
CREATE DATABASE temporal;
CREATE DATABASE temporal_visibility;
