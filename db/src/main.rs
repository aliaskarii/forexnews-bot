use rusqlite::{Connection, Result};
use rusqlite::NO_PARAMS;


fn main() -> Result<()> {
    let conn = Connection::open("news.db")?;
        conn.execute(
        "create table if not exists news (
            id integer primary key,
            title TEXT, 
            country TEXT,
            date DATE ,
            impact TEXT,
            forecast TEXT,
            previous TEXT
         )",
        NO_PARAMS,
    )?;   



    Ok(())
}
