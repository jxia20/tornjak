from flask import Flask, jsonify
from flask_cors import CORS
import sqlite3

app = Flask(__name__)
CORS(app)

def init_db():
    conn = sqlite3.connect('database.db')
    cursor = conn.cursor()
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL
        )
    ''')
    cursor.execute("INSERT INTO users (name) VALUES ('John Doe')")
    cursor.execute("INSERT INTO users (name) VALUES ('Jane Smith')")
    conn.commit()
    conn.close()

@app.route('/api/users', methods=['GET'])
def get_users():
    conn = sqlite3.connect('database.db')
    cursor = conn.cursor()
    cursor.execute('SELECT * FROM users')
    users = cursor.fetchall()
    conn.close()
    return jsonify({'users': [{'id': row[0], 'name': row[1]} for row in users]})

if __name__ == '__main__':
    init_db()
    app.run(port=5000, debug=True)