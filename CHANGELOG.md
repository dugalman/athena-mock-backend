# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.0.] - 2023-06-12

1. Modificar la Entidad Socio: Añadiremos el campo Puntaje.
2. Crear un Programa seeder: Escribiremos un pequeño programa en Go, separado del servidor, cuya única función sea crear los archivos db/ egms.json y db/socios.json con los datos que has especificado.
3. Añadir un Comando al Makefile: Crearemos el comando make seed que ejecutará nuestro programa seeder.
4. Ampliar la API de Socios: Crearemos nuevos endpoints y handlers para consultar y manipular los puntos de un socio.
Añadir Pruebas: Actualizaremos las pruebas de integración para verificar la nueva funcionalidad de puntos.