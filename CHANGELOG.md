# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [4.0.0] - 2023-06-12

¡Excelente! Este es un paso crucial para robustecer la seguridad y la lógica de tu API. Implementar un middleware de autenticación y manejar tokens inválidos es fundamental.

Este escenario es más complejo que los anteriores porque introduce dos conceptos nuevos:
*   Rutas Protegidas: Rutas que requieren un token JWT válido para ser accedidas.
*   Whitelist (Lista Blanca): Una lista de rutas que no requieren autenticación (como /login).

El Plan
1. Refactorizar el LogoutHandler: Lo cambiaremos para que extraiga el userId del token JWT en lugar del cuerpo de la petición. Esto es más seguro y estándar.
2. Actualizar el Middleware de Autenticación: Modificaremos nuestro AuthMiddleware para que sea más robusto y lo aplicaremos a las rutas que necesiten protección.
3. Implementar la Lógica de Whitelist: Modificaremos la configuración de nuestras rutas para aplicar el middleware solo a las rutas que no estén en la lista blanca. gin hace esto muy manejable.
4. Escribir la Prueba de Integración: Crearemos una prueba específica que intente hacer logout con un token inválido/expirado y verifique la respuesta de error 401.


## [3.0.0] - 2023-06-12

1. Definir la Estructura de Respuesta: Crearemos una struct en el paquete model para representar la respuesta JSON.
2. Crear el Handler: Escribiremos el handler que recolecta la información (versión, sistema operativo, etc.) y construye la respuesta.
3. Añadir la Ruta: Registraremos la ruta POST /info en nuestro router.
4. Escribir la Prueba: Añadiremos una nueva prueba de integración para verificar que el endpoint funciona como se espera.

## [2.0.0] - 2023-06-12

1. Modificar la Entidad Socio: Añadiremos el campo Puntaje.
2. Crear un Programa seeder: Escribiremos un pequeño programa en Go, separado del servidor, cuya única función sea crear los archivos db/ egms.json y db/socios.json con los datos que has especificado.
3. Añadir un Comando al Makefile: Crearemos el comando make seed que ejecutará nuestro programa seeder.
4. Ampliar la API de Socios: Crearemos nuevos endpoints y handlers para consultar y manipular los puntos de un socio.
5. Añadir Pruebas: Actualizaremos las pruebas de integración para verificar la nueva funcionalidad de puntos.