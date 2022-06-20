import React, { useState, useEffect } from "react";
import axios from "axios";
import Post from "./post";
import LoggedinNavbar from "../navigation/nav";
import './home.css'

const LoggedInHome = () => {
    const [posts, setPosts] = useState([])

    useEffect(() => {
        axios.get("https://intronus-backend.herokuapp.com/posts/getPosts")
            .then(resp => setPosts(resp.data))
            .catch(err => console.log(err))
    }, [])

    const grid = posts.map(item => {
        return (
            <Post
                key={item}
                value={item}
            />
        )
    })
    return (
        <div className="innerhome">
            <LoggedinNavbar />
            <div>
                <div className="jumbotron p-3 p-md-5 text-white rounded bg-dark">
                    <div className="col-md-6 px-0">
                        <h1 className="display-4 font-italic">Welcome to IntroNus, Batman</h1>
                        <p className="lead my-3">Scroll down around here to find the latest school events</p>
                    </div>
                </div>

                <div className="row mb-2">
                    <div>
                        {grid}
                    </div>
                </div>
            </div>
        </div>
    )
}

export default LoggedInHome
