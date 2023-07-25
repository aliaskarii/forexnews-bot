use rusqlite::{Connection, Result};
use rusqlite::NO_PARAMS;
use std::collections::HashMap;


fn main() -> Result<()> {
    let conn = sqlite::open("news.db").unwrap();
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
