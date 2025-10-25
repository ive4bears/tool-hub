-- Generated SQL schema from ../../hub.db


CREATE TABLE `tools` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` integer,
    `updated_at` integer,
    `name` text,
    `description` text,
    `parameters` text,
    `type` text,
    `log_life_span` text)


CREATE TABLE `concurrency_groups` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` integer,
    `updated_at` integer,
    `name` text,
    `max_concurrent` integer)

CREATE TABLE `command_line_tools` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` integer,
    `updated_at` integer,
    `name` text,
    `description` text,
    `parameters` text,
    `type` text,
    `log_life_span` text,
    `wd` text,
    `cmd` text,
    `env` text,
    `timeout` text,
    `is_stream` numeric,
    `task_queue` text,
    `error` text,
    `status` text,
    `testcases` text,
    `concurrency_group_id` integer,
    CONSTRAINT `fk_command_line_tools_concurrency_group` FOREIGN KEY (`concurrency_group_id`) REFERENCES `concurrency_groups` (`id`))

CREATE TABLE `dependencies` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` integer,
    `updated_at` integer,
    `name` text,
    `description` text,
    `doc` text,
    `url` text,
    `install_cmd` text,
    `testcases` text)

CREATE TABLE `command_line_tool_dependencies` (
    `command_line_tool_id` integer,
    `dependency_id` integer,
    PRIMARY KEY (`command_line_tool_id`, `dependency_id`),
    CONSTRAINT `fk_command_line_tool_dependencies_command_line_tool` FOREIGN KEY (`command_line_tool_id`) REFERENCES `command_line_tools` (`id`),
    CONSTRAINT `fk_command_line_tool_dependencies_dependency` FOREIGN KEY (`dependency_id`) REFERENCES `dependencies` (`id`))

CREATE TABLE `service_tools` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` integer,
    `updated_at` integer,
    `name` text,
    `description` text,
    `parameters` text,
    `type` text,
    `log_life_span` text,
    `start_cmd` text,
    `error` text,
    `status` text,
    `concurrency_group_id` integer,
    CONSTRAINT `fk_service_tools_concurrency_group` FOREIGN KEY (`concurrency_group_id`) REFERENCES `concurrency_groups` (`id`))

CREATE TABLE `http_tools` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` integer,
    `updated_at` integer,
    `name` text,
    `description` text,
    `parameters` text,
    `type` text,
    `log_life_span` text,
    `endpoint` text,
    `method` text,
    `query` text,
    `headers` text,
    `body` text,
    `timeout` text,
    `error` text,
    `status` text,
    `testcases` text,
    `concurrency_group_id` integer,
    CONSTRAINT `fk_http_tools_concurrency_group` FOREIGN KEY (`concurrency_group_id`) REFERENCES `concurrency_groups` (`id`))

CREATE TABLE `calling_logs` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `created_at` integer,
    `updated_at` integer,
    `caller_id` integer,
    `caller_type` text,
    `callee_id` integer,
    `callee_type` text,
    `input` text,
    `output` text,
    `error` text,
    `duration` text)

