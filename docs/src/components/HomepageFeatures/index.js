import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

const FeatureList = [

  {
    title: 'Rapid delivery',
    icon: "⏱️",
    description: (
      <>
        Full REST API wrapper for a SQL table in 5 minutes? No problem. Need customization? We got you covered.
      </>
    ),
  },
  {
    title: 'Concise and ellegant',
    icon: "👔",
    description: (
      <>
        Hate code generation? So do we. Gin REST Framework uses generics to
        provide a concise and ellegant API.
      </>
    ),
  },
  {
    title: 'It\'s Just a library',
    icon: "🧩",
    description: (
      <>
        GRF doesn't enforce any file structure or project layout, you can freely use it with your existing Gin project. Did I mention it's using GORM?
      </>
    ),
  },
];

function Feature({icon, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title} {icon}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
