/* Tema (sistema/chiaro/scuro) e lingua della web UI. Traduzione del DOM: le chiavi sono i testi italiani. */
(function () {
  const T = {
    // navigazione
    'Prenotazioni': ['Bookings', 'Reservas', 'Réservations', 'Buchungen'],
    'Lezioni': ['Classes', 'Clases', 'Cours', 'Kurse'],
    'Profili': ['Profiles', 'Perfiles', 'Profils', 'Profile'],
    'Impostazioni': ['Settings', 'Ajustes', 'Réglages', 'Einstellungen'],
    'Registro': ['Log', 'Registro', 'Journal', 'Protokoll'],
    'Esci': ['Sign out', 'Salir', 'Déconnexion', 'Abmelden'],
    'Sistema': ['System', 'Sistema', 'Système', 'System'], 'Chiaro': ['Light', 'Claro', 'Clair', 'Hell'], 'Scuro': ['Dark', 'Oscuro', 'Sombre', 'Dunkel'],
    'Tema': ['Theme', 'Tema', 'Thème', 'Design'], 'Lingua': ['Language', 'Idioma', 'Langue', 'Sprache'],
    // walkthrough
    'Benvenuto in Wellness Gateway': ['Welcome to Wellness Gateway', 'Bienvenido a Wellness Gateway', 'Bienvenue dans Wellness Gateway', 'Willkommen bei Wellness Gateway'],
    "Questo container prenota da solo le lezioni Technogym mywellness per tutta la famiglia: all'apertura delle prenotazioni e, se una classe è piena, appena si libera un posto (osservazione continua). L'app iPhone Wellness Booking serve per scegliere cosa prenotare e ricevere le notifiche.": ['This container books Technogym mywellness classes for the whole family on its own: when booking opens and, if a class is full, as soon as a spot frees up (continuous watching). The Wellness Booking iPhone app is for choosing what to book and receiving notifications.', 'Este contenedor reserva solo las clases Technogym mywellness para toda la familia: al abrir las reservas y, si una clase está llena, en cuanto se libera una plaza (observación continua). La app de iPhone Wellness Booking sirve para elegir qué reservar y recibir notificaciones.', 'Ce conteneur réserve seul les cours Technogym mywellness pour toute la famille : à l\'ouverture des réservations et, si un cours est complet, dès qu\'une place se libère (observation continue). L\'app iPhone Wellness Booking sert à choisir quoi réserver et recevoir les notifications.', 'Dieser Container bucht die Technogym-mywellness-Kurse für die ganze Familie selbst: bei Öffnung der Buchung und, wenn ein Kurs voll ist, sobald ein Platz frei wird (ständige Beobachtung). Die iPhone-App Wellness Booking dient zur Auswahl und für Mitteilungen.'],
    'Quattro passi: amministratore → profili mywellness → notifiche push → app.': ['Four steps: administrator → mywellness profiles → push notifications → app.', 'Cuatro pasos: administrador → perfiles mywellness → notificaciones push → app.', 'Quatre étapes : administrateur → profils mywellness → notifications push → app.', 'Vier Schritte: Administrator → mywellness-Profile → Push-Mitteilungen → App.'],
    'Inizia': ['Start', 'Empezar', 'Commencer', 'Los geht\'s'],
    '1 · Amministratore': ['1 · Administrator', '1 · Administrador', '1 · Administrateur', '1 · Administrator'],
    "L'utente con cui entrerai nella web UI e nell'app. Potrai aggiungere gli altri familiari dopo.": ['The user you will sign in with on the web UI and in the app. You can add the other family members later.', 'El usuario con el que entrarás en la web y en la app. Podrás añadir a los demás familiares después.', 'L\'utilisateur avec lequel vous vous connecterez à l\'interface web et à l\'app. Vous pourrez ajouter les autres membres ensuite.', 'Der Benutzer, mit dem du dich in der Web-UI und der App anmeldest. Die anderen Familienmitglieder kannst du später hinzufügen.'],
    'Nome utente': ['Username', 'Nombre de usuario', 'Nom d\'utilisateur', 'Benutzername'],
    'Nome visualizzato': ['Display name', 'Nombre visible', 'Nom affiché', 'Anzeigename'],
    'Password (almeno 6 caratteri)': ['Password (at least 6 characters)', 'Contraseña (mínimo 6 caracteres)', 'Mot de passe (au moins 6 caractères)', 'Passwort (mindestens 6 Zeichen)'],
    'Il tuo account mywellness': ['Your mywellness account', 'Tu cuenta mywellness', 'Votre compte mywellness', 'Dein mywellness-Konto'],
    '(opzionale, puoi farlo dopo)': ['(optional, you can do it later)', '(opcional, puedes hacerlo después)', '(facultatif, possible plus tard)', '(optional, geht auch später)'],
    'Se anche tu prenoti al centro, inserisci qui le tue credenziali Technogym mywellness: diventano il tuo profilo personale.': ['If you book at the club too, enter your Technogym mywellness credentials here: they become your personal profile.', 'Si tú también reservas en el centro, introduce aquí tus credenciales Technogym mywellness: serán tu perfil personal.', 'Si vous réservez aussi au club, saisissez ici vos identifiants Technogym mywellness : ils deviennent votre profil personnel.', 'Wenn auch du im Studio buchst, gib hier deine Technogym-mywellness-Zugangsdaten ein: sie werden dein persönliches Profil.'],
    'Email mywellness': ['mywellness email', 'Correo mywellness', 'E-mail mywellness', 'mywellness-E-Mail'],
    'Password mywellness': ['mywellness password', 'Contraseña mywellness', 'Mot de passe mywellness', 'mywellness-Passwort'],
    'Centro (URL widget)': ['Club (widget URL)', 'Centro (URL del widget)', 'Club (URL du widget)', 'Studio (Widget-URL)'],
    'Crea amministratore': ['Create administrator', 'Crear administrador', 'Créer l\'administrateur', 'Administrator anlegen'],
    '2 · Profili mywellness': ['2 · mywellness profiles', '2 · Perfiles mywellness', '2 · Profils mywellness', '2 · mywellness-Profile'],
    'Ogni profilo è un account Technogym mywellness (email e password con cui si prenota al centro). Le password restano cifrate sul NAS; il gateway fa il login e rinnova la sessione da solo. Puoi aggiungerne altri dopo da "Profili".': ['Each profile is a Technogym mywellness account (the email and password used to book at the club). Passwords stay encrypted on the NAS; the gateway logs in and renews the session on its own. You can add more later under "Profiles".', 'Cada perfil es una cuenta Technogym mywellness (correo y contraseña con los que se reserva en el centro). Las contraseñas quedan cifradas en el NAS; el gateway inicia sesión y la renueva solo. Puedes añadir más después en "Perfiles".', 'Chaque profil est un compte Technogym mywellness (e-mail et mot de passe utilisés pour réserver au club). Les mots de passe restent chiffrés sur le NAS ; le gateway se connecte et renouvelle la session seul. Vous pourrez en ajouter d\'autres dans « Profils ».', 'Jedes Profil ist ein Technogym-mywellness-Konto (E-Mail und Passwort fürs Buchen im Studio). Passwörter bleiben verschlüsselt auf dem NAS; das Gateway meldet sich an und erneuert die Sitzung selbst. Weitere kannst du später unter „Profile“ hinzufügen.'],
    'Aggiungi profilo': ['Add profile', 'Añadir perfil', 'Ajouter un profil', 'Profil hinzufügen'],
    'Avanti': ['Next', 'Siguiente', 'Suivant', 'Weiter'],
    '3 · Notifiche push': ['3 · Push notifications', '3 · Notificaciones push', '3 · Notifications push', '3 · Push-Mitteilungen'],
    'La chiave APNs si crea su developer.apple.com → Keys → "+" → Apple Push Notifications service.': ['The APNs key is created at developer.apple.com → Keys → "+" → Apple Push Notifications service.', 'La clave APNs se crea en developer.apple.com → Keys → "+" → Apple Push Notifications service.', 'La clé APNs se crée sur developer.apple.com → Keys → « + » → Apple Push Notifications service.', 'Der APNs-Schlüssel wird unter developer.apple.com → Keys → „+“ → Apple Push Notifications service erstellt.'],
    "4 · Collega l'app": ['4 · Connect the app', '4 · Conecta la app', '4 · Connecter l\'app', '4 · App verbinden'],
    'Indirizzo': ['Address', 'Dirección', 'Adresse', 'Adresse'], 'Utente': ['User', 'Usuario', 'Utilisateur', 'Benutzer'],
    'quello creato al passo 1 (o un familiare)': ['the one created in step 1 (or a family member)', 'el creado en el paso 1 (o un familiar)', 'celui créé à l\'étape 1 (ou un membre de la famille)', 'der in Schritt 1 erstellte (oder ein Familienmitglied)'],
    'Da casa puoi usare anche l\'indirizzo LAN del NAS. Fuori casa serve il tunnel Cloudflare (es. booking.manieridimambro.it).': ['At home you can also use the NAS LAN address. Away from home you need the Cloudflare tunnel (e.g. booking.manieridimambro.it).', 'En casa también puedes usar la dirección LAN del NAS. Fuera de casa hace falta el túnel de Cloudflare (p. ej. booking.manieridimambro.it).', 'À la maison vous pouvez aussi utiliser l\'adresse LAN du NAS. À l\'extérieur il faut le tunnel Cloudflare (ex. booking.manieridimambro.it).', 'Zu Hause kannst du auch die LAN-Adresse des NAS nutzen. Unterwegs brauchst du den Cloudflare-Tunnel (z. B. booking.manieridimambro.it).'],
    'Vai al pannello': ['Go to the dashboard', 'Ir al panel', 'Aller au tableau de bord', 'Zum Dashboard'],
    // login
    'Accedi': ['Sign in', 'Iniciar sesión', 'Connexion', 'Anmelden'], 'Entra': ['Enter', 'Entrar', 'Entrer', 'Los'],
    // profili (form)
    'Etichetta (es. Daniela)': ['Label (e.g. Daniela)', 'Etiqueta (p. ej. Daniela)', 'Libellé (ex. Daniela)', 'Bezeichnung (z. B. Daniela)'],
    'Massimo prenotazioni attive': ['Max active bookings', 'Máximo de reservas activas', 'Réservations actives max', 'Max. aktive Buchungen'],
    'Visibilità': ['Visibility', 'Visibilidad', 'Visibilité', 'Sichtbarkeit'],
    'Famiglia (tutti gli utenti)': ['Family (all users)', 'Familia (todos los usuarios)', 'Famille (tous les utilisateurs)', 'Familie (alle Benutzer)'],
    'Solo io': ['Only me', 'Solo yo', 'Moi seulement', 'Nur ich'],
    'Di chi è': ['Whose is it', 'De quién es', 'À qui est-il', 'Wem gehört es'],
    'È il mio account mywellness': ['It is my mywellness account', 'Es mi cuenta mywellness', 'C\'est mon compte mywellness', 'Es ist mein mywellness-Konto'],
    'Di un familiare senza utente': ['A family member without a user', 'De un familiar sin usuario', 'D\'un membre de la famille sans utilisateur', 'Ein Familienmitglied ohne Benutzer'],
    'Verifica login mywellness…': ['Checking mywellness login…', 'Verificando el login de mywellness…', 'Vérification de la connexion mywellness…', 'Prüfe mywellness-Anmeldung…'],
    'Profilo aggiunto e verificato.': ['Profile added and verified.', 'Perfil añadido y verificado.', 'Profil ajouté et vérifié.', 'Profil hinzugefügt und geprüft.'],
    'Profili mywellness': ['mywellness profiles', 'Perfiles mywellness', 'Profils mywellness', 'mywellness-Profile'],
    'Aggiungi e verifica': ['Add and verify', 'Añadir y verificar', 'Ajouter et vérifier', 'Hinzufügen und prüfen'],
    'Etichetta': ['Label', 'Etiqueta', 'Libellé', 'Bezeichnung'], 'Account': ['Account', 'Cuenta', 'Compte', 'Konto'], 'Centro': ['Club', 'Centro', 'Club', 'Studio'],
    'Max': ['Max', 'Máx', 'Max', 'Max'], 'Attive': ['Active', 'Activas', 'Actives', 'Aktiv'], 'Login': ['Login', 'Login', 'Connexion', 'Login'],
    'Rifai login': ['Log in again', 'Repetir login', 'Se reconnecter', 'Erneut anmelden'], 'Rimuovi': ['Remove', 'Eliminar', 'Supprimer', 'Entfernen'],
    'Nessun profilo.': ['No profiles.', 'Ningún perfil.', 'Aucun profil.', 'Keine Profile.'],
    'Utenti del gateway': ['Gateway users', 'Usuarios del gateway', 'Utilisateurs du gateway', 'Gateway-Benutzer'],
    "Ogni utente entra nell'app con il suo nome utente. Se ha un account mywellness, inseriscilo qui sotto: diventa il suo profilo personale (visibile alla famiglia).": ['Each user signs in to the app with their own username. If they have a mywellness account, enter it below: it becomes their personal profile (visible to the family).', 'Cada usuario entra en la app con su nombre de usuario. Si tiene cuenta mywellness, introdúcela abajo: será su perfil personal (visible para la familia).', 'Chaque utilisateur se connecte à l\'app avec son nom d\'utilisateur. S\'il a un compte mywellness, saisissez-le ci-dessous : il devient son profil personnel (visible par la famille).', 'Jeder Benutzer meldet sich in der App mit seinem Benutzernamen an. Hat er ein mywellness-Konto, trag es unten ein: es wird sein persönliches Profil (für die Familie sichtbar).'],
    'Nome': ['Name', 'Nombre', 'Nom', 'Name'], 'Password gateway': ['Gateway password', 'Contraseña del gateway', 'Mot de passe du gateway', 'Gateway-Passwort'],
    'Ruolo': ['Role', 'Rol', 'Rôle', 'Rolle'], 'Amministratore': ['Administrator', 'Administrador', 'Administrateur', 'Administrator'],
    'Email mywellness (opzionale)': ['mywellness email (optional)', 'Correo mywellness (opcional)', 'E-mail mywellness (facultatif)', 'mywellness-E-Mail (optional)'],
    'Aggiungi utente': ['Add user', 'Añadir usuario', 'Ajouter un utilisateur', 'Benutzer hinzufügen'],
    'Elimina': ['Delete', 'Eliminar', 'Supprimer', 'Löschen'], 'admin': ['admin', 'admin', 'admin', 'Admin'], 'utente': ['user', 'usuario', 'utilisateur', 'Benutzer'],
    'Creazione in corso…': ['Creating…', 'Creando…', 'Création…', 'Wird erstellt…'],
    // prenotazioni
    'Aggiorna': ['Refresh', 'Actualizar', 'Actualiser', 'Aktualisieren'],
    'Lezioni seguite': ['Tracked classes', 'Clases seguidas', 'Cours suivis', 'Verfolgte Kurse'],
    'Prenotate su mywellness': ['Booked on mywellness', 'Reservadas en mywellness', 'Réservés sur mywellness', 'Auf mywellness gebucht'],
    "Tutte le prenotazioni attive del profilo, fatte dal gateway o dall'app/sito Technogym. Da qui puoi disdire.": ['All active bookings of the profile, made by the gateway or in the Technogym app/website. You can cancel from here.', 'Todas las reservas activas del perfil, hechas por el gateway o en la app/web de Technogym. Desde aquí puedes anular.', 'Toutes les réservations actives du profil, faites par le gateway ou dans l\'app/le site Technogym. Vous pouvez annuler d\'ici.', 'Alle aktiven Buchungen des Profils, vom Gateway oder in der Technogym-App/Website gemacht. Hier kannst du stornieren.'],
    'Ultime attività': ['Latest activity', 'Últimas actividades', 'Dernières activités', 'Letzte Aktivitäten'],
    'Registro completo →': ['Full log →', 'Registro completo →', 'Journal complet →', 'Vollständiges Protokoll →'],
    'Nessuna lezione seguita: vai in Lezioni.': ['No tracked classes: go to Classes.', 'Ninguna clase seguida: ve a Clases.', 'Aucun cours suivi : allez dans Cours.', 'Keine verfolgten Kurse: gehe zu Kurse.'],
    'Nessuna prenotazione attiva.': ['No active bookings.', 'Ninguna reserva activa.', 'Aucune réservation active.', 'Keine aktiven Buchungen.'],
    'Nessuna attività per questo profilo.': ['No activity for this profile.', 'Sin actividad para este perfil.', 'Aucune activité pour ce profil.', 'Keine Aktivität für dieses Profil.'],
    'Prenotata dal gateway': ['Booked by the gateway', 'Reservada por el gateway', 'Réservé par le gateway', 'Vom Gateway gebucht'],
    'Prenotata da mywellness (app/web)': ['Booked on mywellness (app/web)', 'Reservada en mywellness (app/web)', 'Réservé sur mywellness (app/web)', 'Auf mywellness gebucht (App/Web)'],
    'Disdici': ['Cancel', 'Anular', 'Annuler', 'Stornieren'], 'Riprova': ['Retry', 'Reintentar', 'Réessayer', 'Erneut'], 'Stop ricorrenza': ['Stop recurrence', 'Detener repetición', 'Arrêter la récurrence', 'Wiederholung beenden'],
    'Disdire la prenotazione su mywellness?': ['Cancel the booking on mywellness?', '¿Anular la reserva en mywellness?', 'Annuler la réservation sur mywellness ?', 'Buchung auf mywellness stornieren?'],
    'Disdetta inviata.': ['Cancellation sent.', 'Anulación enviada.', 'Annulation envoyée.', 'Stornierung gesendet.'],
    'In attesa': ['Waiting', 'En espera', 'En attente', 'Wartet'], 'Prenotazione in corso': ['Booking in progress', 'Reserva en curso', 'Réservation en cours', 'Buchung läuft'],
    'Osservazione: piena': ['Watching: full', 'Observación: llena', 'Observation : complet', 'Beobachtung: voll'], "Lista d'attesa": ['Waiting list', 'Lista de espera', 'Liste d\'attente', 'Warteliste'],
    'Prenotata': ['Booked', 'Reservada', 'Réservé', 'Gebucht'], 'Errore': ['Error', 'Error', 'Erreur', 'Fehler'], 'Scaduta': ['Expired', 'Caducada', 'Expirée', 'Abgelaufen'], 'Disdetta': ['Cancelled', 'Anulada', 'Annulée', 'Storniert'],
    // lezioni
    'Cerca lezione, istruttore, sala': ['Search class, trainer, room', 'Buscar clase, instructor, sala', 'Rechercher cours, coach, salle', 'Kurs, Trainer, Raum suchen'],
    'Prenota questa': ['Book this', 'Reservar esta', 'Réserver celui-ci', 'Diesen buchen'], 'Ogni settimana': ['Every week', 'Cada semana', 'Chaque semaine', 'Jede Woche'],
    "In lista d'attesa": ['On the waiting list', 'En lista de espera', 'En liste d\'attente', 'Auf der Warteliste'], 'Non prenotabile online': ['Not bookable online', 'No reservable en línea', 'Non réservable en ligne', 'Nicht online buchbar'],
    "Piena · lista d'attesa": ['Full · waiting list', 'Llena · lista de espera', 'Complet · liste d\'attente', 'Voll · Warteliste'],
    'Aggiunta alle prenotazioni automatiche.': ['Added to automatic booking.', 'Añadida a las reservas automáticas.', 'Ajouté aux réservations automatiques.', 'Zur automatischen Buchung hinzugefügt.'],
    'Nessuna lezione.': ['No classes.', 'Ninguna clase.', 'Aucun cours.', 'Keine Kurse.'],
    // impostazioni
    "Segui l'orario di apertura comunicato dal centro": ['Follow the opening time reported by the club', 'Seguir la hora de apertura indicada por el centro', 'Suivre l\'heure d\'ouverture indiquée par le club', 'Der vom Studio gemeldeten Öffnungszeit folgen'],
    'Regole di prenotazione (quanti giorni prima apre ogni tipo di lezione)': ['Booking rules (how many days before each kind of class opens)', 'Reglas de reserva (cuántos días antes abre cada tipo de clase)', 'Règles de réservation (combien de jours avant chaque type de cours ouvre)', 'Buchungsregeln (wie viele Tage vorher jede Kursart öffnet)'],
    'Testo nel nome': ['Text in the name', 'Texto en el nombre', 'Texte dans le nom', 'Text im Namen'], 'Giorni prima': ['Days before', 'Días antes', 'Jours avant', 'Tage vorher'], 'Ora': ['Time', 'Hora', 'Heure', 'Uhrzeit'],
    'Lezioni intercettate': ['Matched classes', 'Clases coincidentes', 'Cours concernés', 'Erfasste Kurse'], 'tutte le altre': ['all the others', 'todas las demás', 'tous les autres', 'alle anderen'],
    'nessuna lezione in calendario': ['no class in the schedule', 'ninguna clase en el calendario', 'aucun cours au planning', 'kein Kurs im Kalender'], 'scrivi un testo': ['type a text', 'escribe un texto', 'saisissez un texte', 'Text eingeben'],
    '+ Aggiungi regola': ['+ Add rule', '+ Añadir regla', '+ Ajouter une règle', '+ Regel hinzufügen'],
    'Anticipo (ms)': ['Lead time (ms)', 'Antelación (ms)', 'Avance (ms)', 'Vorlauf (ms)'], "Insisti dopo l'apertura (s)": ['Keep trying after opening (s)', 'Insistir tras la apertura (s)', 'Insister après l\'ouverture (s)', 'Nach Öffnung weiterversuchen (s)'],
    'Osservazione: controlla ogni (s)': ['Watching: check every (s)', 'Observación: comprobar cada (s)', 'Observation : vérifier toutes les (s)', 'Beobachtung: prüfen alle (s)'], 'Giorni di calendario': ['Calendar days', 'Días de calendario', 'Jours de calendrier', 'Kalendertage'],
    'Notifiche prioritarie (Time Sensitive)': ['Time Sensitive notifications', 'Notificaciones prioritarias (Time Sensitive)', 'Notifications prioritaires (Time Sensitive)', 'Zeitkritische Mitteilungen (Time Sensitive)'],
    'Salva': ['Save', 'Guardar', 'Enregistrer', 'Sichern'], 'Impostazioni salvate.': ['Settings saved.', 'Ajustes guardados.', 'Réglages enregistrés.', 'Einstellungen gesichert.'],
    "Solo l'amministratore può modificare le impostazioni.": ['Only the administrator can change the settings.', 'Solo el administrador puede cambiar los ajustes.', 'Seul l\'administrateur peut modifier les réglages.', 'Nur der Administrator kann die Einstellungen ändern.'],
    'Stato': ['Status', 'Estado', 'État', 'Status'], 'Versione': ['Version', 'Versión', 'Version', 'Version'], 'Avviato': ['Started', 'Iniciado', 'Démarré', 'Gestartet'],
    'Push APNs': ['APNs push', 'Push APNs', 'Push APNs', 'APNs-Push'], 'attivo': ['active', 'activo', 'actif', 'aktiv'], 'non configurato': ['not configured', 'no configurado', 'non configuré', 'nicht konfiguriert'],
    'Indirizzo pubblico': ['Public address', 'Dirección pública', 'Adresse publique', 'Öffentliche Adresse'],
    'Invia notifica di prova ai miei dispositivi': ['Send a test notification to my devices', 'Enviar notificación de prueba a mis dispositivos', 'Envoyer une notification test à mes appareils', 'Testmitteilung an meine Geräte senden'],
    'APNs non configurato': ['APNs not configured', 'APNs no configurado', 'APNs non configuré', 'APNs nicht konfiguriert'],
    // registro
    'Registro attività': ['Activity log', 'Registro de actividad', 'Journal d\'activité', 'Aktivitätsprotokoll'], 'Tutti i profili': ['All profiles', 'Todos los perfiles', 'Tous les profils', 'Alle Profile'],
    'Tutti i livelli': ['All levels', 'Todos los niveles', 'Tous les niveaux', 'Alle Stufen'], 'Successi': ['Successes', 'Éxitos', 'Succès', 'Erfolge'], 'Avvisi': ['Warnings', 'Avisos', 'Avertissements', 'Warnungen'], 'Errori': ['Errors', 'Errores', 'Erreurs', 'Fehler'],
    'Nessuna attività.': ['No activity.', 'Sin actividad.', 'Aucune activité.', 'Keine Aktivität.'],
    'Si aggiorna da solo ogni 20 secondi': ['Refreshes on its own every 20 seconds', 'Se actualiza solo cada 20 segundos', 'S\'actualise seul toutes les 20 secondes', 'Aktualisiert sich alle 20 Sekunden selbst'],
    'Account mywellness': ['mywellness account', 'Cuenta mywellness', 'Compte mywellness', 'mywellness-Konto'],
    'Collega account mywellness': ['Link mywellness account', 'Vincular cuenta mywellness', 'Lier un compte mywellness', 'mywellness-Konto verknüpfen'],
    'Collega': ['Link', 'Vincular', 'Lier', 'Verknüpfen'], 'Annulla': ['Cancel', 'Cancelar', 'Annuler', 'Abbrechen'], 'Modifica': ['Edit', 'Editar', 'Modifier', 'Bearbeiten'],
    'Nuova password gateway (vuoto = invariata)': ['New gateway password (empty = unchanged)', 'Nueva contraseña del gateway (vacío = sin cambios)', 'Nouveau mot de passe du gateway (vide = inchangé)', 'Neues Gateway-Passwort (leer = unverändert)'],
    'Non hai ancora collegato il tuo account mywellness: usa "Aggiungi profilo" con "È il mio account mywellness" (oppure "Collega account mywellness" nella tabella utenti).': ['You have not linked your mywellness account yet: use "Add profile" with "It is my mywellness account" (or "Link mywellness account" in the users table).', 'Aún no has vinculado tu cuenta mywellness: usa "Añadir perfil" con "Es mi cuenta mywellness" (o "Vincular cuenta mywellness" en la tabla de usuarios).', 'Vous n\'avez pas encore lié votre compte mywellness : utilisez « Ajouter un profil » avec « C\'est mon compte mywellness » (ou « Lier un compte mywellness » dans la table des utilisateurs).', 'Du hast dein mywellness-Konto noch nicht verknüpft: nutze „Profil hinzufügen“ mit „Es ist mein mywellness-Konto“ (oder „mywellness-Konto verknüpfen“ in der Benutzertabelle).'],
    'Utenti e account mywellness': ['Users and mywellness accounts', 'Usuarios y cuentas mywellness', 'Utilisateurs et comptes mywellness', 'Benutzer und mywellness-Konten'],
    "Ogni utente del gateway è una persona con il suo account Technogym mywellness: entra nell'app con nome utente e password del gateway, mentre il gateway usa l'account mywellness per prenotare. I profili \"famiglia\" sono visibili a tutti; ognuno vede di default il proprio.": ['Each gateway user is a person with their own Technogym mywellness account: they sign in to the app with the gateway username and password, while the gateway uses the mywellness account to book. "Family" profiles are visible to everyone; each person sees their own by default.', 'Cada usuario del gateway es una persona con su cuenta Technogym mywellness: entra en la app con usuario y contraseña del gateway, mientras el gateway usa la cuenta mywellness para reservar. Los perfiles "familia" son visibles para todos; cada uno ve el suyo por defecto.', 'Chaque utilisateur du gateway est une personne avec son compte Technogym mywellness : il se connecte à l\'app avec l\'identifiant et le mot de passe du gateway, tandis que le gateway réserve avec le compte mywellness. Les profils « famille » sont visibles par tous ; chacun voit le sien par défaut.', 'Jeder Gateway-Benutzer ist eine Person mit eigenem Technogym-mywellness-Konto: er meldet sich in der App mit Gateway-Benutzername und -Passwort an, während das Gateway mit dem mywellness-Konto bucht. „Familien“-Profile sehen alle; jeder sieht standardmäßig sein eigenes.'],
    'Nome utente (per entrare)': ['Username (to sign in)', 'Nombre de usuario (para entrar)', 'Nom d\'utilisateur (pour se connecter)', 'Benutzername (zum Anmelden)'],
    'Password gateway (almeno 6 caratteri)': ['Gateway password (at least 6 characters)', 'Contraseña del gateway (mínimo 6 caracteres)', 'Mot de passe du gateway (au moins 6 caractères)', 'Gateway-Passwort (mindestens 6 Zeichen)'],
    'Visibilità del profilo': ['Profile visibility', 'Visibilidad del perfil', 'Visibilité du profil', 'Sichtbarkeit des Profils'],
    'Solo questa persona': ['Only this person', 'Solo esta persona', 'Cette personne seulement', 'Nur diese Person'],
    'Solo il proprietario': ['Owner only', 'Solo el propietario', 'Propriétaire seulement', 'Nur Besitzer'],
    'Appartiene a': ['Belongs to', 'Pertenece a', 'Appartient à', 'Gehört zu'], '— nessun utente —': ['— no user —', '— ningún usuario —', '— aucun utilisateur —', '— kein Benutzer —'],
    'senza utente': ['no user', 'sin usuario', 'sans utilisateur', 'ohne Benutzer'],
    'Nuova password mywellness (vuoto = invariata)': ['New mywellness password (empty = unchanged)', 'Nueva contraseña mywellness (vacío = sin cambios)', 'Nouveau mot de passe mywellness (vide = inchangé)', 'Neues mywellness-Passwort (leer = unverändert)'],
    'Nessun account mywellness: aggiungi un utente qui sopra.': ['No mywellness account: add a user above.', 'Ninguna cuenta mywellness: añade un usuario arriba.', 'Aucun compte mywellness : ajoutez un utilisateur ci-dessus.', 'Kein mywellness-Konto: füge oben einen Benutzer hinzu.'],
    'Email e password mywellness sono obbligatorie: ogni utente è una persona con il suo account mywellness.': ['mywellness email and password are required: each user is a person with their own mywellness account.', 'El correo y la contraseña de mywellness son obligatorios: cada usuario es una persona con su cuenta mywellness.', 'E-mail et mot de passe mywellness sont obligatoires : chaque utilisateur est une personne avec son compte mywellness.', 'mywellness-E-Mail und -Passwort sind Pflicht: jeder Benutzer ist eine Person mit eigenem mywellness-Konto.'],
    'Utente creato, ma profilo mywellness non aggiunto:': ['User created, but the mywellness profile was not added:', 'Usuario creado, pero el perfil mywellness no se añadió:', 'Utilisateur créé, mais le profil mywellness n\'a pas été ajouté :', 'Benutzer erstellt, aber mywellness-Profil nicht hinzugefügt:'],
    'Utente e profilo mywellness creati.': ['User and mywellness profile created.', 'Usuario y perfil mywellness creados.', 'Utilisateur et profil mywellness créés.', 'Benutzer und mywellness-Profil erstellt.'],
    "Esci dalla lista d'attesa": ['Leave the waiting list', 'Salir de la lista de espera', 'Quitter la liste d\'attente', 'Warteliste verlassen'],
    "Uscire dalla lista d'attesa su mywellness? La lezione non verrà più seguita.": ['Leave the waiting list on mywellness? The class will no longer be tracked.', '¿Salir de la lista de espera en mywellness? La clase dejará de seguirse.', 'Quitter la liste d\'attente sur mywellness ? Le cours ne sera plus suivi.', 'Warteliste auf mywellness verlassen? Der Kurs wird nicht mehr verfolgt.'],
    "Uscita dalla lista d'attesa.": ['Left the waiting list.', 'Has salido de la lista de espera.', 'Liste d\'attente quittée.', 'Warteliste verlassen.'],
    '+ Nuovo utente': ['+ New user', '+ Nuevo usuario', '+ Nouvel utilisateur', '+ Neuer Benutzer'], 'Nuovo utente': ['New user', 'Nuevo usuario', 'Nouvel utilisateur', 'Neuer Benutzer'],
    'Tocca un utente per modificare tutti i suoi parametri.': ['Click a user to edit all their settings.', 'Toca un usuario para modificar todos sus parámetros.', 'Cliquez sur un utilisateur pour modifier tous ses paramètres.', 'Klicke auf einen Benutzer, um alle Einstellungen zu bearbeiten.'],
    'Account mywellness senza utente': ['mywellness accounts without a user', 'Cuentas mywellness sin usuario', 'Comptes mywellness sans utilisateur', 'mywellness-Konten ohne Benutzer'],
    "Account rimasti senza persona (per esempio dopo l'eliminazione di un utente): assegnali a un utente dalla sua finestra, oppure rimuovili.": ['Accounts left without a person (for example after deleting a user): assign them to a user from their dialog, or remove them.', 'Cuentas sin persona (por ejemplo tras eliminar un usuario): asígnalas a un usuario desde su ventana, o elimínalas.', 'Comptes sans personne (par exemple après la suppression d\'un utilisateur) : attribuez-les à un utilisateur depuis sa fenêtre, ou supprimez-les.', 'Konten ohne Person (z. B. nach dem Löschen eines Benutzers): weise sie einem Benutzer in dessen Fenster zu oder entferne sie.'],
    'Nessuno.': ['None.', 'Ninguna.', 'Aucun.', 'Keine.'], 'nessuno': ['none', 'ninguna', 'aucun', 'keines'],
    'Login mywellness': ['mywellness login', 'Login mywellness', 'Connexion mywellness', 'mywellness-Login'],
    'Accesso al gateway': ['Gateway access', 'Acceso al gateway', 'Accès au gateway', 'Gateway-Zugang'],
    'Obbligatorio: ogni utente è una persona con il suo account mywellness.': ['Required: each user is a person with their own mywellness account.', 'Obligatorio: cada usuario es una persona con su cuenta mywellness.', 'Obligatoire : chaque utilisateur est une personne avec son compte mywellness.', 'Pflicht: jeder Benutzer ist eine Person mit eigenem mywellness-Konto.'],
    'Nessun account collegato.': ['No account linked.', 'Ninguna cuenta vinculada.', 'Aucun compte lié.', 'Kein Konto verknüpft.'],
    'Assegna un account esistente': ['Assign an existing account', 'Asignar una cuenta existente', 'Attribuer un compte existant', 'Vorhandenes Konto zuweisen'],
    '— nuovo account qui sotto —': ['— new account below —', '— nueva cuenta abajo —', '— nouveau compte ci-dessous —', '— neues Konto unten —'],
    'Rifai login mywellness': ['Log in to mywellness again', 'Repetir login mywellness', 'Se reconnecter à mywellness', 'Erneut bei mywellness anmelden'],
    'Scollega account': ['Unlink account', 'Desvincular cuenta', 'Dissocier le compte', 'Konto trennen'], 'Rimuovi account': ['Remove account', 'Eliminar cuenta', 'Supprimer le compte', 'Konto entfernen'],
    'Crea utente': ['Create user', 'Crear usuario', 'Créer l\'utilisateur', 'Benutzer erstellen'], 'Chiudi': ['Close', 'Cerrar', 'Fermer', 'Schließen'], 'Elimina utente': ['Delete user', 'Eliminar usuario', 'Supprimer l\'utilisateur', 'Benutzer löschen'],
    'Salvataggio…': ['Saving…', 'Guardando…', 'Enregistrement…', 'Speichern…'], 'Login mywellness riuscito.': ['mywellness login succeeded.', 'Login mywellness correcto.', 'Connexion mywellness réussie.', 'mywellness-Login erfolgreich.'],
    "Per cambiare la tua password usa l'app o chiedi a un amministratore.": ['To change your password use the app or ask an administrator.', 'Para cambiar tu contraseña usa la app o pide a un administrador.', 'Pour changer votre mot de passe, utilisez l\'app ou demandez à un administrateur.', 'Zum Ändern deines Passworts nutze die App oder frage einen Administrator.'],
    'Obbligatorio per gli utenti normali (senza account non vedrebbero nulla); facoltativo per un amministratore solo locale, che vede e gestisce tutti i profili.': ['Required for regular users (without an account they would see nothing); optional for a local-only administrator, who sees and manages every profile.', 'Obligatorio para los usuarios normales (sin cuenta no verían nada); opcional para un administrador solo local, que ve y gestiona todos los perfiles.', 'Obligatoire pour les utilisateurs normaux (sans compte ils ne verraient rien) ; facultatif pour un administrateur local, qui voit et gère tous les profils.', 'Pflicht für normale Benutzer (ohne Konto sähen sie nichts); optional für einen rein lokalen Administrator, der alle Profile sieht und verwaltet.'],
    'Un utente normale deve avere il suo account mywellness (altrimenti non vedrebbe nulla). Per un account solo locale scegli il ruolo Amministratore.': ['A regular user must have their own mywellness account (otherwise they would see nothing). For a local-only account choose the Administrator role.', 'Un usuario normal debe tener su cuenta mywellness (si no, no vería nada). Para una cuenta solo local elige el rol Administrador.', 'Un utilisateur normal doit avoir son compte mywellness (sinon il ne verrait rien). Pour un compte local, choisissez le rôle Administrateur.', 'Ein normaler Benutzer braucht sein mywellness-Konto (sonst sähe er nichts). Für ein rein lokales Konto wähle die Rolle Administrator.'],
    "Ogni utente entra nell'app con nome utente e password del gateway. Un utente normale ha il suo account Technogym mywellness e vede solo quello; un amministratore vede e gestisce tutti i profili e può anche essere solo locale, senza account mywellness. Tocca un utente per modificare tutti i suoi parametri.": ['Each user signs in to the app with the gateway username and password. A regular user has their own Technogym mywellness account and sees only that; an administrator sees and manages every profile and may be local-only, without a mywellness account. Click a user to edit all their settings.', 'Cada usuario entra en la app con usuario y contraseña del gateway. Un usuario normal tiene su cuenta Technogym mywellness y ve solo esa; un administrador ve y gestiona todos los perfiles y puede ser solo local, sin cuenta mywellness. Toca un usuario para modificar todos sus parámetros.', 'Chaque utilisateur se connecte à l\'app avec l\'identifiant et le mot de passe du gateway. Un utilisateur normal a son compte Technogym mywellness et ne voit que celui-ci ; un administrateur voit et gère tous les profils et peut être local uniquement, sans compte mywellness. Cliquez sur un utilisateur pour modifier tous ses paramètres.', 'Jeder Benutzer meldet sich mit Gateway-Benutzername und -Passwort an. Ein normaler Benutzer hat sein Technogym-mywellness-Konto und sieht nur dieses; ein Administrator sieht und verwaltet alle Profile und kann rein lokal ohne mywellness-Konto sein. Klicke auf einen Benutzer, um alle Einstellungen zu bearbeiten.'],
    'Nessun profilo': ['No profiles', 'Ningún perfil', 'Aucun profil', 'Keine Profile'],
    'Rimuovere il profilo e le sue lezioni seguite?': ['Remove the profile and its tracked classes?', '¿Eliminar el perfil y sus clases seguidas?', 'Supprimer le profil et ses cours suivis ?', 'Profil und seine verfolgten Kurse entfernen?'],
    'Eliminare utente?': ['Delete user?', '¿Eliminar usuario?', 'Supprimer l\'utilisateur ?', 'Benutzer löschen?'],
  };
  // frasi con numeri/date: regex → template
  const P = [
    [/^Prenotazioni attive (\d+)\/(\d+) · motore attivo, prossimo controllo (.+)$/, ['Active bookings $1/$2 · engine running, next check $3', 'Reservas activas $1/$2 · motor activo, próxima comprobación $3', 'Réservations actives $1/$2 · moteur actif, prochaine vérification $3', 'Aktive Buchungen $1/$2 · Motor aktiv, nächste Prüfung $3']],
    [/^(\d+) liberi su (\d+)$/, ['$1 free of $2', '$1 libres de $2', '$1 libres sur $2', '$1 frei von $2']],
    [/^Apre (.+)$/, ['Opens $1', 'Abre $1', 'Ouvre $1', 'Öffnet $1']],
    [/^Prenoto (.+)$/, ['Booking $1', 'Reservaré $1', 'Je réserve $1', 'Buche $1']],
    [/^Ultimo controllo (.+) · tentativi (\d+)$/, ['Last check $1 · attempts $2', 'Última comprobación $1 · intentos $2', 'Dernière vérification $1 · tentatives $2', 'Letzte Prüfung $1 · Versuche $2']],
    [/^(\d+) in corso su (\d+)$/, ['$1 in progress of $2', '$1 en curso de $2', '$1 en cours sur $2', '$1 laufend von $2']],
    [/^profilo di (.+)$/, ['profile of $1', 'perfil de $1', 'profil de $1', 'Profil von $1']],
    [/^ok (.+)$/, ['ok $1', 'ok $1', 'ok $1', 'ok $1']],
    [/^inviata a (\d+) dispositivi$/, ['sent to $1 devices', 'enviada a $1 dispositivos', 'envoyée à $1 appareils', 'an $1 Geräte gesendet']],
  ];
  const LANGS = ['en', 'es', 'fr', 'de'];
  let lang = 'it';
  try { lang = localStorage.getItem('wg.lang') || (navigator.language || 'it').slice(0, 2); } catch (e) {}
  if (!['it', ...LANGS].includes(lang)) lang = 'en';

  function tr(text) {
    if (lang === 'it') return text;
    const i = LANGS.indexOf(lang);
    const key = text.trim();
    if (!key) return text;
    if (T[key]) return text.replace(key, T[key][i]);
    for (const [re, out] of P) { const m = key.match(re); if (m) return text.replace(key, key.replace(re, out[i])); }
    return text;
  }
  const SKIP = new Set(['SCRIPT', 'STYLE', 'CODE']);
  function walk(node) {
    if (node.nodeType === 3) { const v = tr(node.nodeValue); if (v !== node.nodeValue) node.nodeValue = v; return; }
    if (node.nodeType !== 1 || SKIP.has(node.tagName)) return;
    for (const a of ['placeholder', 'title']) if (node.hasAttribute && node.hasAttribute(a)) { const v = tr(node.getAttribute(a)); if (v !== node.getAttribute(a)) node.setAttribute(a, v); }
    if (node.tagName === 'INPUT' && node.type === 'button') { const v = tr(node.value); if (v !== node.value) node.value = v; }
    node.childNodes.forEach(walk);
  }
  function apply(root) { if (lang !== 'it') walk(root || document.body); document.documentElement.lang = lang; }
  window.i18n = { get lang() { return lang; }, apply, t: tr };
  T['Mostra password'] = ['Show password', 'Mostrar contraseña', 'Afficher le mot de passe', 'Passwort anzeigen'];
  T['Nascondi password'] = ['Hide password', 'Ocultar contraseña', 'Masquer le mot de passe', 'Passwort verbergen'];
  T['Aspetto e lingua'] = ['Appearance and language', 'Aspecto e idioma', 'Apparence et langue', 'Darstellung und Sprache'];
  T['Valgono per questo browser.'] = ['They apply to this browser.', 'Valen para este navegador.', 'Ils s\'appliquent à ce navigateur.', 'Gelten für diesen Browser.'];

  function bindPrefs() {
    const th = document.getElementById('theme'), lg = document.getElementById('lang');
    if (!th || !lg) return;
    try { th.value = localStorage.getItem('wg.theme') || 'system'; } catch (e) {}
    lg.value = lang;
    th.onchange = () => { try { localStorage.setItem('wg.theme', th.value); } catch (e) {} if (th.value === 'system') delete document.documentElement.dataset.theme; else document.documentElement.dataset.theme = th.value; };
    lg.onchange = () => { try { localStorage.setItem('wg.lang', lg.value); } catch (e) {} location.reload(); };
  }
  /* occhio per mostrare/nascondere le password */
  function addEyes(root) {
    (root.querySelectorAll ? root.querySelectorAll('input[type=password]:not([data-eye])') : []).forEach(inp => {
      inp.dataset.eye = '1';
      const wrap = document.createElement('div'); wrap.className = 'pw';
      inp.parentNode.insertBefore(wrap, inp); wrap.appendChild(inp);
      const EYE = '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>';
      const EYE_OFF = '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>';
      const b = document.createElement('button'); b.type = 'button'; b.className = 'eye'; b.title = tr('Mostra password'); b.innerHTML = EYE; b.setAttribute('aria-label', tr('Mostra password'));
      b.onclick = () => { const show = inp.type === 'password'; inp.type = show ? 'text' : 'password'; b.innerHTML = show ? EYE_OFF : EYE; b.title = tr(show ? 'Nascondi password' : 'Mostra password'); };
      wrap.appendChild(b);
    });
  }
  window.i18n = Object.assign(window.i18n || {}, { bindPrefs, addEyes });

  document.addEventListener('DOMContentLoaded', () => {
    bindPrefs(); addEyes(document.body);
    apply(document.body);
    new MutationObserver(muts => { for (const m of muts) m.addedNodes.forEach(n => { if (n.nodeType === 1) addEyes(n); }); })
      .observe(document.body, { childList: true, subtree: true });
    new MutationObserver(muts => { if (lang === 'it') return; for (const m of muts) { m.addedNodes.forEach(walk); if (m.type === 'characterData') walk(m.target); } })
      .observe(document.body, { childList: true, subtree: true, characterData: true });
  });
})();
