import React from 'react';
import DinoGame from 'react-chrome-dino-ts';
import 'react-chrome-dino-ts/index.css';
import '../styles/NotFound.css';

const NotFound = () => {
  return (
    <div className="not-found-container">
      <h1>404 - 页面未找到</h1>
      <p>哎呀！您要查找的页面不存在。不过别灰心，玩个游戏放松一下吧。</p>
      <div className="dino-game-container">
        <DinoGame />
      </div>
    </div>
  );
};

export default NotFound;