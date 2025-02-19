import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

const FeatureList = [

  {
    title: 'Rapid prototyping',
    icon: "⏱️",
    description: (
      <>
        Complete REST API resource over SQL table in several minutes. Need to plug-in custom business logic? No problem.
      </>
    ),
  },
  {
    title: 'Concise and ellegant',
    icon: "👔",
    description: (
      <>
        Hate code generation? So do we. GRF uses generics to hide the boring stuff and let you focus on what brings value to your project.

      </>
    ),
  },
  {
    title: 'It\'s Just a library',
    icon: "🧩",
    description: (
      <>
        GRF doesn't enforce any file structure or project layout, you can freely use it with your existing Gin project. Uses pluggable storage layer, with GORM as the default option.
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
