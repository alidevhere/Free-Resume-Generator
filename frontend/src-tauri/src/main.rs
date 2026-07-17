use std::fs;
use std::process::{Child, Command, Stdio};
use std::sync::Mutex;

use tauri::{AppHandle, Manager, RunEvent};

struct BackendProcess(Mutex<Option<Child>>);

fn backend_binary_name() -> &'static str {
    if cfg!(target_os = "windows") {
        "resume-generator.exe"
    } else {
        "resume-generator"
    }
}

fn backend_resource_dir(app: &AppHandle) -> Result<std::path::PathBuf, String> {
    let resource_dir = app.path().resource_dir().map_err(|e| e.to_string())?;
    let bundled_backend_dir = resource_dir.join("backend");

    if bundled_backend_dir.exists() {
        return Ok(bundled_backend_dir);
    }

    // In `tauri dev`, resources may not be copied into `target/debug`.
    let dev_backend_dir = std::path::PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("resources/backend");
    if dev_backend_dir.exists() {
        return Ok(dev_backend_dir);
    }

    Ok(bundled_backend_dir)
}

fn spawn_backend(app: &AppHandle) -> Result<Child, String> {
    let backend_dir = backend_resource_dir(app)?;
    let binary_path = backend_dir.join(backend_binary_name());

    if !binary_path.exists() {
        return Err(format!("backend sidecar binary not found at {}", binary_path.display()));
    }

    let app_data_dir = app.path().app_data_dir().map_err(|e| e.to_string())?;
    fs::create_dir_all(&app_data_dir).map_err(|e| format!("create app data dir: {e}"))?;

    let db_path = app_data_dir.join("resumes.db");
    let seed_path = backend_dir.join("resume.json");

    Command::new(binary_path)
        .arg("-server")
        .arg(":8080")
        .arg("-db")
        .arg(db_path)
        .arg("-seed")
        .arg(seed_path)
        .current_dir(&backend_dir)
        .stdout(Stdio::inherit())
        .stderr(Stdio::inherit())
        .spawn()
        .map_err(|e| format!("failed to start backend: {e}"))
}

fn stop_backend(app: &AppHandle) {
    if let Some(state) = app.try_state::<BackendProcess>() {
        if let Ok(mut guard) = state.0.lock() {
            if let Some(mut child) = guard.take() {
                let _ = child.kill();
                let _ = child.wait();
            }
        }
    }
}

fn main() {
    tauri::Builder::default()
        .setup(|app| {
            let child = spawn_backend(&app.handle())
                .map_err(|e| -> Box<dyn std::error::Error> { e.into() })?;
            app.manage(BackendProcess(Mutex::new(Some(child))));
            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("error while building tauri application")
        .run(|app, event| {
            if let RunEvent::Exit = event {
                stop_backend(app);
            }
        });
}
